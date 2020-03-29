package cache

import (
	"sync"
)

type cond func(interface{}, interface{}) bool

// Cache implements a thread safe cache
type Cache struct {
	items map[interface{}]interface{}
	lock  sync.RWMutex // the lock for reading or writing
}

// NewCache constructs an cache instance
func NewCache() (*Cache, error) {
	c := &Cache{
		items: make(map[interface{}]interface{}),
	}
	return c, nil
}

// Purge is used to completely clear the cache
func (c *Cache) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for k := range c.items {
		delete(c.items, k)
	}
}

// Add adds a value to the cache, return false if the entry already exists in cache and update the value
func (c *Cache) Add(key, value interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Check for existing item
	if _, ok := c.items[key]; ok {
		c.items[key] = value

		return false
	}

	c.items[key] = value

	return true
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if ent, ok := c.items[key]; ok {
		return ent, true
	}
	return nil, false
}

// Contains checks if a key is in the cache, without return the value related to the key
func (c *Cache) Contains(key interface{}) (ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, ok = c.items[key]
	return ok
}

// Remove removes the provided key from the cache, return true if the
// key was contained.
func (c *Cache) Remove(key interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.items[key]; ok {
		delete(c.items, key)
		return true
	}
	return false
}

// Keys returns a slice of the keys in the cache, out of the order
func (c *Cache) Keys() []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	keys := make([]interface{}, len(c.items))
	i := 0
	for key := range c.items {
		keys[i] = key
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.items)
}

// RemoveWithCond removes elements if their keys are equal to the
// provided key according to fn
func (c *Cache) RemoveWithCond(key interface{}, fn cond) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	flag := false
	for iterKey := range c.items {
		if fn(key, iterKey) {
			delete(c.items, iterKey)
			flag = true
		}
	}
	return flag
}
