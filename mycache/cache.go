package mycache

import (
	"mycache/lru"
	"sync"
)

/*是对 lru.Cache 的封装
* 添加了锁来保证并发操作的安全性*/
type cache struct {
	mu       sync.Mutex
	lru      *lru.Cache
	maxBytes int64
}

// 添加 k-v 对
func (cache *cache) put(key string, value ByteView) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.lru == nil {
		cache.lru = lru.NewCache(cache.maxBytes)
	}
	cache.lru.Put(key, value)
}

// 获取值
func (cache *cache) get(key string) (value ByteView, ok bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.lru == nil {
		return
	}
	if v, ok := cache.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}

// 删除 k-v 对
func (cache *cache) delete(key string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.lru != nil {
		cache.lru.Delete(key)
	}
}

// 判断 k-v 对是否存在
func (cache *cache) exists(key string) bool {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.lru != nil {
		return false
	}
	return cache.lru.IsExists(key)
}
