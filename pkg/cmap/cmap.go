package cmap

import "sync"

// CMap is a thread-safe (concurrent) implementation of map[string]int
// protected by a sync.RWMutex for safe concurrent access.
type CMap struct {
	mu sync.RWMutex
	m  map[string]int
}

// New creates and returns a new initialized CMap instance.
// The returned object is ready to use.
func New() *CMap {
	return &CMap{
		m: make(map[string]int),
	}
}

// Get returns the value associated with the key and a boolean indicating existence.
// Thread-safe read operation.
func (c *CMap) Get(key string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.m[key]
	return val, ok
}

// Set updates or creates a key-value pair in the map.
// Thread-safe write operation.
func (c *CMap) Set(key string, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
}

// Delete removes a key from the map. No-op if key doesn't exist.
// Thread-safe write operation.
func (c *CMap) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, key)
}

// Len returns the current number of elements in the map.
// The count reflects the state at the moment of calling.
func (c *CMap) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.m)
}

// Iterate applies function f to all key-value pairs sequentially.
// Iteration is performed under a read-lock, therefore:
// - Order of iteration is not guaranteed (same as native Go map)
// - Function f MUST NOT modify the map (may cause deadlock)
// - Operation is safe for concurrent access
func (c *CMap) Iterate(f func(key string, value int)) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for k, v := range c.m {
		f(k, v)
	}
}

// Eq compares the internal map state with the provided map[string]int.
// Returns true if both maps have identical key-value pairs.
func (c *CMap) Eq(compareWith map[string]int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Fast path for different sizes
	if len(c.m) != len(compareWith) {
		return false
	}

	// Compare all entries
	for k, v := range c.m {
		if cmpVal, ok := compareWith[k]; !ok || cmpVal != v {
			return false
		}
	}

	return true
}

// Clear removes all elements from the map.
// Thread-safe write operation.
func (c *CMap) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = make(map[string]int)
}

// Add increments the value for a key by specified delta.
// Thread-safe write operation. If key doesn't exist, initializes it with delta.
func (c *CMap) Add(key string, delta int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] += delta
}
