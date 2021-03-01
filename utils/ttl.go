package utils

import (
	"sync"
	"time"
)

// https://github.com/Konstantin8105/SimpleTTL
// entry - typical element of cache
type entry struct {
	value  interface{}
	expiry *time.Time
}

// Cache - simple implementation of cache
// More information: https://en.wikipedia.org/wiki/Time_to_live
type Cache struct {
	timeTTL time.Duration
	cache   map[string]*entry
	lock    *sync.RWMutex
}

// NewCache - initialization of new cache.
// For avoid mistake - minimal time to live is 1 minute.
// For simplification, - key is string and cache haven`t stop method
func NewCache(interval time.Duration) *Cache {
	if interval < time.Second {
		interval = time.Second
	}
	cache := &Cache{
		timeTTL: interval,
		cache:   make(map[string]*entry),
		lock:    &sync.RWMutex{},
	}
	go func() {
		ticker := time.NewTicker(cache.timeTTL)
		for {
			// wait of ticker
			now := <-ticker.C

			// remove entry outside TTL
			cache.lock.Lock()
			for id, entry := range cache.cache {
				if entry.expiry != nil && entry.expiry.Before(now) {
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

	if ok && e.expiry != nil && e.expiry.After(time.Now()) {
		return e.value, true
	}
	return nil, false
}

func (cache *Cache) GetAndUpdate(key string, ttl time.Duration) (interface{}, bool) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	e, ok := cache.cache[key]

	if ok && e.expiry != nil {
		expiry := time.Now().Add(ttl)
		e.expiry = &expiry
		return e.value, true
	}
	return nil, false
}

// Add - add key/value in cache
func (cache *Cache) Add(key string, value interface{}, ttl time.Duration) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	expiry := time.Now().Add(ttl)

	cache.cache[key] = &entry{
		value:  value,
		expiry: &expiry,
	}
}

// GetKeys - return all keys of cache map
func (cache *Cache) GetKeys() []interface{} {
	cache.lock.RLock()
	defer cache.lock.RUnlock()

	keys := make([]interface{}, len(cache.cache))
	var i int
	for k := range cache.cache {
		keys[i] = k
		i++
	}
	return keys
}
