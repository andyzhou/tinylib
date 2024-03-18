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
	DefaultRingReplicas = 3
)

type Hash func(data []byte) uint32

//hash ring face
type HashRing struct {
	hash     Hash
	replicas int
	ring     []int // Sorted
	hashMap  map[int]string
	nodes map[string]bool
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

// Add adds some nodes to the hash.
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
}