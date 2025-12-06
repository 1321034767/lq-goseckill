package auth

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// ConsistentHashRing 简单的一致性哈希实现，用于分布式鉴权节点选择
type ConsistentHashRing struct {
	hash       func(data []byte) uint32
	replicas   int
	keys       []int // 已排序的虚拟节点哈希
	hashMap    map[int]string
	mu         sync.RWMutex
	nodeLookup map[string]struct{}
}

// NewConsistentHashRing 创建哈希环，nodes为空时会生成一个默认节点，避免空指针
func NewConsistentHashRing(nodes []string, replicas int) *ConsistentHashRing {
	if replicas <= 0 {
		replicas = 50
	}
	if len(nodes) == 0 {
		nodes = []string{"auth-node-default"}
	}
	ch := &ConsistentHashRing{
		hash:       crc32.ChecksumIEEE,
		replicas:   replicas,
		hashMap:    make(map[int]string),
		nodeLookup: make(map[string]struct{}),
	}
	ch.Add(nodes...)
	return ch
}

// Add 批量添加节点
func (c *ConsistentHashRing) Add(nodes ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, node := range nodes {
		if _, exists := c.nodeLookup[node]; exists {
			continue
		}
		c.nodeLookup[node] = struct{}{}
		for i := 0; i < c.replicas; i++ {
			hash := int(c.hash([]byte(node + "#" + strconv.Itoa(i))))
			c.keys = append(c.keys, hash)
			c.hashMap[hash] = node
		}
	}
	sort.Ints(c.keys)
}

// GetNode 根据 key 获取负责的节点
func (c *ConsistentHashRing) GetNode(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.keys) == 0 {
		return ""
	}
	hash := int(c.hash([]byte(key)))
	// 二分查找
	idx := sort.Search(len(c.keys), func(i int) bool { return c.keys[i] >= hash })
	if idx == len(c.keys) {
		idx = 0
	}
	return c.hashMap[c.keys[idx]]
}
