package git4go

import (
	"sync"
)

type Cache struct {
	enabled        bool
	currentStorage int
	maxStorage     int
	maxObjectSize  map[ObjectType]int
	lock           sync.RWMutex
	cache          map[Oid]Object
	usedMemory     int
}

func NewCache() *Cache {
	return &Cache{
		enabled:    true,
		maxStorage: 256 * 1024 * 1024,
		maxObjectSize: map[ObjectType]int{
			ObjectCommit: 4096,
			ObjectTree:   4096,
			ObjectBlob:   0,
			ObjectTag:    4096,
		},
		cache: make(map[Oid]Object),
	}
}

func (c *Cache) Get(oid *Oid) Object {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.cache[*oid]
}

func (c *Cache) Set(obj Object) {
	size := obj.Size()
	if !c.enabled || c.maxObjectSize[obj.Type()] < size {
		return
	}
	c.lock.Lock()
	defer c.lock.Unlock()

	c.cache[*obj.Id()] = obj
	c.usedMemory += size
}
