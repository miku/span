package crossref

import "sync"

// Cache for int keys and string values with a thread-safe setter.
type IntStringCache struct {
	mu      *sync.RWMutex
	Entries map[int]string
}

// NewMemberCache creates a new in memory members cache.
func NewIntStringCache() IntStringCache {
	return IntStringCache{Entries: make(map[int]string), mu: new(sync.RWMutex)}
}

// Set sets the string value for an int key, threadsafe.
func (c *IntStringCache) Set(k int, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Entries[k] = v
}
