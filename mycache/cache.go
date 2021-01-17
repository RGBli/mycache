package mycache

import (
	"mycache/lru"
	"sync"
)

type cache struct {
	mu          sync.Mutex
	lru         *lru.Cache
	maxCapacity int
}

func (cache *cache) add(key, value string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.lru == nil {
		cache.lru = lru.NewCache(cache.maxCapacity)
	}
	cache.lru.Put(key, value)
}

func (cache *cache) get(key string) (value string, ok bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.lru == nil {
		return
	}
	return cache.lru.Get(key)
}
