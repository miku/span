package crossref

import "sync"

// Cache for int keys and string values, thread-safe.
type IntStringCache struct {
	lock    *sync.RWMutex
	Entries map[int]string
}

// NewMemberCache creates a new in memory members cache.
func NewIntStringCache() IntStringCache {
	return IntStringCache{Entries: make(map[int]string), lock: new(sync.RWMutex)}
}

func (c *IntStringCache) Set(k int, v string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Entries[k] = v
}
