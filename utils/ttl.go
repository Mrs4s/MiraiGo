package utils

import (
	"sync"
	"time"
)

// https://github.com/Konstantin8105/SimpleTTL
// entry - typical element of cache
type entry struct {
	expiry time.Time
	value  interface{}
}

// Cache - simple implementation of cache
// More information: https://en.wikipedia.org/wiki/Time_to_live
type Cache struct {
	lock  sync.RWMutex
	cache map[string]*entry
}

// NewCache - initialization of new cache.
// For avoid mistake - minimal time to live is 1 minute.
// For simplification, - key is string and cache haven`t stop method
func NewCache(interval time.Duration) *Cache {
	if interval < time.Second {
		interval = time.Second
	}
	cache := &Cache{cache: make(map[string]*entry)}
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
func (cache *Cache) Count() int {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	return len(cache.cache)
}

// Get - return value from cache
func (cache *Cache) Get(key string) (interface{}, bool) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	e, ok := cache.cache[key]

	if ok && e.expiry.After(time.Now()) {
		return e.value, true
	}
	return nil, false
}

func (cache *Cache) GetAndUpdate(key string, ttl time.Duration) (interface{}, bool) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	if e, ok := cache.cache[key]; ok {
		e.expiry = time.Now().Add(ttl)
		return e.value, true
	}
	return nil, false
}

// Add - add key/value in cache
func (cache *Cache) Add(key string, value interface{}, ttl time.Duration) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	cache.cache[key] = &entry{
		value:  value,
		expiry: time.Now().Add(ttl),
	}
}

// GetKeys - return all keys of cache map
func (cache *Cache) GetKeys() []string {
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
