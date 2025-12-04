package common

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type units []uint32

func (x units) Len() int { return len(x) }

func (x units) Less(i, j int) bool { return x[i] < x[j] }

func (x units) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

// empty error
var errEmpty = errors.New("consistent: empty circle")

type Consistent struct {
	// hash circle
	circle map[uint32]string
	// hash circle sorted
	sortedHashes units
	// virtual node number
	VirtualNode int
	// rw mutex
	sync.RWMutex
}

func NewConsistent() *Consistent {
	return &Consistent{
		circle:      make(map[uint32]string),
		VirtualNode: 20,
	}
}

func (c *Consistent) generateKey(element string, index int) string {
	return element + strconv.Itoa(index)
}

func (c *Consistent) hashKey(key string) uint32 {
	// use crc32
	return crc32.ChecksumIEEE([]byte(key))
}

func (c *Consistent) updateSortedHashes() {
	hashes := c.sortedHashes[:0]
	if cap(c.sortedHashes)/(c.VirtualNode*4) > len(c.circle) {
		hashes = nil
	}
	for k := range c.circle {
		hashes = append(hashes, k)
	}
	// sort
	sort.Sort(hashes)
	c.sortedHashes = hashes
}

func (c *Consistent) Add(element string) {
	c.Lock()
	defer c.Unlock()
	c.add(element)
}

func (c *Consistent) add(element string) {
	for i := 0; i < c.VirtualNode; i++ {
		c.circle[c.hashKey(c.generateKey(element, i))] = element
	}
	c.updateSortedHashes()
}

func (c *Consistent) Remove(element string) {
	c.Lock()
	defer c.Unlock()
	c.remove(element)
}

func (c *Consistent) remove(element string) {
	for i := 0; i < c.VirtualNode; i++ {
		delete(c.circle, c.hashKey(c.generateKey(element, i)))
	}
	c.updateSortedHashes()
}

func (c *Consistent) Get(name string) (string, error) {
	c.RLock()
	defer c.RUnlock()
	if len(c.circle) == 0 {
		return "", errEmpty
	}
	hash := c.hashKey(name)
	i := c.search(hash)
	return c.circle[c.sortedHashes[i]], nil
}

func (c *Consistent) search(hash uint32) int {
	f := func(x int) bool {
		return c.sortedHashes[x] > hash
	}
	i := sort.Search(c.sortedHashes.Len(), f)
	if i >= c.sortedHashes.Len() {
		i = 0
	}
	return i
}
