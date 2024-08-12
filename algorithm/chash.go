package algorithm

/*
 * consistent hash algorithm
 * base on `https://github.com/golang/groupcache/consistenthash`
 */

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

//inter macro define
const (
	DefaultRingReplicas = 32
)

type Hash func(data []byte) uint32

//hash ring face
type HashRing struct {
	hash     Hash
	replicas int
	ring     []int // Sorted
	hashMap  map[int]string
	nodes    map[string]bool
	sync.RWMutex
}

//construct
func NewHashRingDefault() *HashRing {
	hf := func(key []byte) uint32 {
		i, err := strconv.ParseInt(string(key), 10, 64)
		if err != nil {
			panic(any(err))
		}
		return uint32(i)
	}
	return NewHashRing(DefaultRingReplicas, hf)
}
func NewHashRing(replicas int, fn Hash) *HashRing {
	m := &HashRing{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
		nodes: map[string]bool{},
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// IsEmpty returns true if there are no items available.
func (m *HashRing) IsEmpty() bool {
	return len(m.ring) == 0
}

//get ring length
func (m *HashRing) GetRingLen() int {
	return len(m.ring)/m.replicas
}

//get all rings
func (m *HashRing) GetRings() []string {
	ringLen := m.GetRingLen()
	if ringLen <= 0 {
		return nil
	}
	result := make([]string, 0)
	for i := 0; i < ringLen; i++ {
		ring := m.GetByIdx(i)
		result = append(result, ring)
	}
	return result
}

//get idx by ring
//idx value range 0 ~ N
func (m *HashRing) GetIdxByRing(ring string) int {
	if ring == "" {
		return -1
	}
	ringLen := m.GetRingLen()
	if ringLen <= 0 {
		return -1
	}
	for i := 0; i < ringLen; i++ {
		tmpRing := m.GetByIdx(i)
		if tmpRing == ring {
			//found it
			return i
		}
	}
	return 0
}

//get first ring
func (m *HashRing) GetFirstRing() string {
	return m.GetByIdx(0)
}

//get next ring
func (m *HashRing) GetNextRing(ring string) string {
	if ring == "" {
		return ""
	}

	//get ring len
	ringLen := m.GetRingLen()
	if ringLen <= 0 {
		return ""
	}

	//get ring idx
	ringIdx := m.GetIdxByRing(ring)
	if ringIdx < 0 || ringIdx >= (ringLen - 1) {
		return ""
	}

	//get next ring
	nextRingIdx := ringIdx + 1
	return m.GetByIdx(nextRingIdx)
}

//get ring by idx
func (m *HashRing) GetByIdx(idx int) string {
	//check
	if idx < 0 {
		return ""
	}
	ringLen := len(m.ring)
	if ringLen <= 0 || idx >= ringLen {
		return ""
	}
	m.Lock()
	defer m.Unlock()
	return m.hashMap[m.ring[idx]]
}

// Get gets the closest item in the hash to the provided key.
func (m *HashRing) Get(key string) string {
	//if no hash ring, return empty
	if m.IsEmpty() {
		return ""
	}

	//calculate hash value
	hash := int(m.hash([]byte(key)))

	// Binary search for appropriate replica.
	idx := sort.Search(len(m.ring), func(i int) bool { return m.ring[i] >= hash })

	// Means we have cycled back to the first replica.
	if idx == len(m.ring) {
		idx = 0
	}

	//return real node of hash value
	m.Lock()
	defer m.Unlock()
	return m.hashMap[m.ring[idx]]
	//return m.hashMap[m.ring[idx%len(m.ring)]]
}

//get all
func (m *HashRing) GetAll() map[int]string {
	m.Lock()
	defer m.Unlock()
	return m.hashMap
}

// Add adds some nodes to the hash.
//node value should be numeric string format
//like "1", "2", etc.
func (m *HashRing) Add(nodes ...string) {
	m.Lock()
	defer m.Unlock()
	for _, node := range nodes {
		//check node
		_, ok := m.nodes[node]
		if ok {
			continue
		}
		m.nodes[node] = true

		//create batch replicas node
		for i := 0; i < m.replicas; i++ {
			//cal virtual hash value
			hash := int(m.hash([]byte(strconv.Itoa(i) + node)))
			//add into hash ring
			m.ring = append(m.ring, hash)
			//map the hash value and node info
			m.hashMap[hash] = node
		}
	}
	//sort hash ring
	sort.Ints(m.ring)
}