package util

import (
	"fmt"
	"hash/fnv"
	"log"
	"sync"
	"time"
	"unsafe"
)

/*
 * shared map face
 * - multi shared maps to resolve map memory
 */

//default setup
const (
	defaultMapShardCount            = 16
	defaultMapShardRebuildThreshold = 1024 * 1024
	defaultMapShardCheckInternal    = 10 * time.Second
)

//config
type SharedMapConf struct {
	ShardCount       uint
	RebuildThreshold int
	CheckInternal    time.Duration
}

//single shard map
type shard struct {
	items map[int64]interface{} //id -> data
	sync.RWMutex
}

//face info
type ShardedMap struct {
	conf   *SharedMapConf
	shards map[uint]*shard //sub shard container
}

//construct
func NewMapShard(configs ...*SharedMapConf) *ShardedMap {
	//setup default config
	conf := &SharedMapConf{
		ShardCount: defaultMapShardCount,
		RebuildThreshold: defaultMapShardRebuildThreshold,
		CheckInternal: defaultMapShardCheckInternal,
	}

	//use input config
	if len(configs) > 0 {
		conf = configs[0]
	}

	//self init
	this := &ShardedMap{
		conf: conf,
		shards: map[uint]*shard{},
	}

	//inter init
	this.interInit()
	return this
}

//set key and value
func (f *ShardedMap) Set(key int64, value interface{}) {
	s := f.getShard(key)
	s.Lock()
	defer s.Unlock()
	s.items[key] = value
}

//get key value
func (f *ShardedMap) Get(key int64) (interface{}, bool) {
	s := f.getShard(key)
	s.RLock()
	defer s.Unlock()
	v, ok := s.items[key]
	return v, ok
}

//delete key
func (f *ShardedMap) Delete(key int64) {
	s := f.getShard(key)
	s.Lock()
	defer s.Unlock()
	delete(s.items, key)
}

//run sub function on all shards
func (f *ShardedMap) Range(sf func(key int64, value interface{})) {
	for _, s := range f.shards {
		s.RLock()
		for k, v := range s.items {
			sf(k, v)
		}
		s.RUnlock()
	}
}

//force rebuild to release memory by sys
func (f *ShardedMap) ForceRebuild() {
	for _, s := range f.shards {
		s.rebuild()
	}
}

//get single shard map items
func (f *ShardedMap) GetShardItems(idx uint) map[int64]interface{} {
	shardObj, ok := f.shards[idx]
	if ok && shardObj != nil {
		return shardObj.items
	}
	return nil
}

//get config
func (f *ShardedMap) GetConf() *SharedMapConf {
	return f.conf
}

//get target shard
func (f *ShardedMap) getShard(key int64) *shard {
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%d", key)))
	return f.shards[uint(h.Sum32())%f.conf.ShardCount]
}

//sub shard map check
func (f *ShardedMap) periodicCheck() {
	//init ticker
	ticker := time.NewTicker(f.conf.CheckInternal)
	defer ticker.Stop()

	//loop ticker
	for range ticker.C {
		for i, s := range f.shards {
			mem := s.estimateMemory()
			count := 0
			s.RLock()
			count = len(s.items)
			s.RUnlock()
			if mem > f.conf.RebuildThreshold {
				log.Printf("Shard %d: items=%d, est_mem=%d bytes\n", i, count, mem)
				log.Printf("Shard %d memory %d exceeds threshold, rebuilding...\n", i, mem)
				s.rebuild()
			}
		}
	}
}

//inter init
func (f *ShardedMap) interInit() {
	//init sub shard map
	for i := uint(0); i < f.conf.ShardCount; i++ {
		f.shards[i] = newShard()
	}

	//start detect process
	go f.periodicCheck()
}

////////////////////
//api of sub shard
////////////////////
func newShard() *shard {
	return &shard{
		items: make(map[int64]interface{}),
	}
}

//calculate shard memory usage
func (s *shard) estimateMemory() int {
	s.RLock()
	defer s.RUnlock()
	size := 0
	for k, v := range s.items {
		size += int(unsafe.Sizeof(k)) + int(unsafe.Sizeof(v))
	}
	return size
}

//rebuild shard, move active data to new map
func (s *shard) rebuild() {
	s.Lock()
	defer s.Unlock()
	newMap := make(map[int64]interface{}, len(s.items))
	for k, v := range s.items {
		newMap[k] = v
	}
	s.items = newMap
}