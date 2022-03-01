package utils

import (
	"sync"
	"time"
)

// https://github.com/Konstantin8105/SimpleTTL
// entry - typical element of cache
type entry[T any] struct {
	expiry time.Time
	value  T
}

// Cache - simple implementation of cache
// More information: https://en.wikipedia.org/wiki/Time_to_live
type Cache[T any] struct {
	lock  sync.RWMutex
	cache map[string]*entry[T]
}

// NewCache - initialization of new cache.
// For avoid mistake - minimal time to live is 1 minute.
// For simplification, - key is string and cache haven`t stop method
func NewCache[T any](interval time.Duration) *Cache[T] {
	if interval < time.Second {
		interval = time.Second
	}
	cache := &Cache[T]{cache: make(map[string]*entry[T])}
	go func() {
		ticker := time.NewTicker(interval)
		for {
			// wait of ticker
			now := <-ticker.C

			// remove entry outside TTL
			cache.lock.Lock()
			for id, entry := range cache.cache {
				if entry == nil || entry.expiry.Before(now) {
					delete(cache.cache, id)
				}
			}
			cache.lock.Unlock()
		}
	}()
	return cache
}

// Count - return amount element of TTL map.
func (cache *Cache[_]) Count() int {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	return len(cache.cache)
}

// Get - return value from cache
func (cache *Cache[_]) Get(key string) (interface{}, bool) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	e, ok := cache.cache[key]

	if ok && e.expiry.After(time.Now()) {
		return e.value, true
	}
	return nil, false
}

func (cache *Cache[_]) GetAndUpdate(key string, ttl time.Duration) (interface{}, bool) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	if e, ok := cache.cache[key]; ok {
		e.expiry = time.Now().Add(ttl)
		return e.value, true
	}
	return nil, false
}

// Add - add key/value in cache
func (cache *Cache[T]) Add(key string, value T, ttl time.Duration) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	cache.cache[key] = &entry[T]{
		value:  value,
		expiry: time.Now().Add(ttl),
	}
}

// GetKeys - return all keys of cache map
func (cache *Cache[T]) GetKeys() []string {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	keys := make([]string, len(cache.cache))
	var i int
	for k := range cache.cache {
		keys[i] = k
		i++
	}
	return keys
}
