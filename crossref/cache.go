package crossref

import "sync"

// IntStringCache for int keys and string values with a thread-safe setter.
// TODO(miku): move to something more generic
type IntStringCache struct {
	mu      *sync.RWMutex
	Entries map[int]string
}

// NewIntStringCache creates a new in memory members cache.
func NewIntStringCache() IntStringCache {
	return IntStringCache{Entries: make(map[int]string), mu: new(sync.RWMutex)}
}

// Set sets the string value for an int key, threadsafe.
func (c *IntStringCache) Set(k int, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Entries[k] = v
}
