package tun

import (
	"sync"
	"time"
)

type cacheItem[T any] struct {
	expires time.Time
	data    T
}

type cache[K comparable, V any] struct {
	sync.RWMutex
	cache map[K]*cacheItem[V]
}

func newCache[K comparable, V any]() *cache[K, V] {
	return &cache[K, V]{
		cache: make(map[K]*cacheItem[V]),
	}
}

func (c *cache[K, V]) Check() {
	c.Lock()
	c.UnsafeCheck()
	c.Unlock()
}

func (c *cache[K, V]) Contains(key K) bool {
	c.RLock()
	_, ok := c.cache[key]
	c.RUnlock()
	return ok
}

func (c *cache[K, V]) Load(key K) (value V, ok bool) {
	c.RLock()
	value, ok = c.UnsafeLoad(key)
	c.RUnlock()
	return
}

func (c *cache[K, V]) Store(key K, value V, ttl time.Duration) {
	c.Lock()
	c.UnsafeStore(key, value, ttl)
	c.Unlock()
}

func (c *cache[K, V]) UnsafeCheck() {
	now := time.Now()
	for k, v := range c.cache {
		if now.After(v.expires) {
			delete(c.cache, k)
		}
	}
}

func (c *cache[K, V]) UnsafeContains(key K) bool {
	_, ok := c.cache[key]
	return ok
}

func (c *cache[K, V]) UnsafeLoad(key K) (value V, ok bool) {
	var item *cacheItem[V]
	item, ok = c.cache[key]
	if ok {
		value = item.data
	}
	return
}

func (c *cache[K, V]) UnsafeStore(key K, value V, ttl time.Duration) {
	c.cache[key] = &cacheItem[V]{
		expires: time.Now().Add(ttl),
		data:    value,
	}
}
