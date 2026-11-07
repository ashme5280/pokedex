package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	mu       sync.Mutex
	cacheMap map[string]cacheEntry
}

type cacheEntry struct {
	createdAt 	time.Time
	val 		[]byte
}

func NewCache() Cache {
	return Cache{
		cacheMap: make(map[string]cacheEntry),
	}
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cacheMap[key] = cacheEntry{
		createdAt: time.Now(),
		val:	val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.cacheMap[key]
	if !exists {
		return nil, false
	}
	return entry.val, true
}
