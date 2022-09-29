package cache

import (
	"github.com/zyedidia/generic/cache"
	"sync"
)

type mutexCache[K comparable, V any] struct {
	*cache.Cache[K, V]
	mutex sync.Mutex
}

func newMutexCache[K comparable, V any](c *cache.Cache[K, V]) *mutexCache[K, V] {
	return &mutexCache[K, V]{
		Cache: c,
	}
}

func (c *mutexCache[K, V]) Get(key K) (V, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.Cache.Get(key)
}

func (c *mutexCache[K, V]) Put(key K, value V) {
	c.mutex.Lock()
	c.Cache.Put(key, value)
	c.mutex.Unlock()
}

func (c *mutexCache[K, V]) Remove(key K) {
	c.mutex.Lock()
	c.Cache.Remove(key)
	c.mutex.Unlock()
}

func (c *mutexCache[K, V]) Resize(capacity int) {
	c.mutex.Lock()
	c.Cache.Resize(capacity)
	c.mutex.Unlock()
}

func (c *mutexCache[K, V]) Size() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.Cache.Size()
}

func (c *mutexCache[K, V]) Capacity() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.Cache.Capacity()
}

func (c *mutexCache[K, V]) Each(fn func(key K, value V)) {
	c.mutex.Lock()
	c.Cache.Each(fn)
	c.mutex.Unlock()
}

func (c *mutexCache[K, V]) SetEvictCallback(fn func(key K, value V)) {
	c.mutex.Lock()
	c.Cache.SetEvictCallback(fn)
	c.mutex.Unlock()
}
