package algorithm

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

//inter macro define
const (
	DefaultVirtualNodeCount = 100
)

//face info
type XConsistent struct {
	//排序的hash虚拟结点
	hashSortedNodes []uint32

	//虚拟结点对应的结点信息
	circle map[uint32]string

	//已绑定的结点
	nodes map[string]bool
	orgNodes []string

	//虚拟结点数
	virtualNodeCount int

	//map读写锁
	sync.RWMutex
}

//construct
func NewXConsistent() *XConsistent {
	this := &XConsistent{
		hashSortedNodes: []uint32{},
		circle: map[uint32]string{},
		nodes: map[string]bool{},
		orgNodes:[]string{},
	}
	return this
}

//add node
func (c *XConsistent) Add(node string, virtualNodeCounts ...int) error {
	var (
		virtualNodeCount int
	)
	if node == "" {
		return nil
	}
	if virtualNodeCounts != nil && len(virtualNodeCounts) > 0 {
		virtualNodeCount = virtualNodeCounts[0]
	}
	if virtualNodeCount <= 0 {
		virtualNodeCount = DefaultVirtualNodeCount
	}

	c.Lock()
	defer c.Unlock()

	if c.circle == nil {
		c.circle = map[uint32]string{}
	}
	if c.nodes == nil {
		c.nodes = map[string]bool{}
	}

	if _, ok := c.nodes[node]; ok {
		return errors.New("node already existed")
	}
	c.nodes[node] = true
	c.orgNodes = append(c.orgNodes, node)

	//增加虚拟结点
	for i := 0; i < virtualNodeCount; i++ {
		virtualKey := c.hashKey(node + strconv.Itoa(i))
		c.circle[virtualKey] = node
		c.hashSortedNodes = append(c.hashSortedNodes, virtualKey)
	}

	//虚拟结点排序
	sort.Slice(c.hashSortedNodes, func(i, j int) bool {
		return c.hashSortedNodes[i] < c.hashSortedNodes[j]
	})

	return nil
}

//get key node
func (c *XConsistent) GetNode(key string) string {
	c.RLock()
	defer c.RUnlock()

	hash := c.hashKey(key)
	i := c.getPosition(hash)

	return c.circle[c.hashSortedNodes[i]]
}

//get orgin nodes
func (c *XConsistent) GetOrgNodes() []string {
	return c.orgNodes
}

//////////////
//private func
//////////////

func (c *XConsistent) hashKey(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

func (c *XConsistent) getPosition(hash uint32) int {
	i := sort.Search(len(c.hashSortedNodes), func(i int) bool { return c.hashSortedNodes[i] >= hash })

	if i < len(c.hashSortedNodes) {
		if i == len(c.hashSortedNodes)-1 {
			return 0
		} else {
			return i
		}
	} else {
		return len(c.hashSortedNodes) - 1
	}
}