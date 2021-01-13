package lru

import (
	"container/list"
)

// Cache 是一个 lru 的 cache
type Cache struct {
	currCapacity int
	maxCapacity  int
	cache        map[string]*list.Element
	list         *list.List
}

type entry struct {
	key   string
	value string
}

// NewCache 用于创建一个 Cache 实例
func NewCache(maxCapacity int) *Cache {
	if maxCapacity > 0 {
		return &Cache{
			maxCapacity: maxCapacity,
			cache:       make(map[string]*list.Element),
			list:        list.New(),
		}
	}
	return nil
}

// Put 用于添加一个 key-value 对到 Cache 中
func (lruCache *Cache) Put(key, value string) {
	if ele, ok := lruCache.cache[key]; ok {
		lruCache.list.MoveToFront(ele)
		kv := lruCache.cache[key].Value.(*entry)
		kv.value = value
	} else {
		ele := lruCache.list.PushFront(&entry{key: key, value: value})
		lruCache.cache[key] = ele
		lruCache.currCapacity++
	}
	for lruCache.currCapacity > lruCache.maxCapacity && lruCache.maxCapacity > 0 {
		lruCache.RemoveOldest()
	}
}

// Get 用于从 Cache 中获取指定 key 的 value
func (lruCache *Cache) Get(key string) (value string, ok bool) {
	if ele, ok := lruCache.cache[key]; ok {
		return ele.Value.(*entry).value, ok
	}
	return
}

// RemoveOldest 用于淘汰 Cache 中的元素
func (lruCache *Cache) RemoveOldest() {
	ele := lruCache.list.Back()
	if ele != nil {
		lruCache.list.Remove(ele)
		key := ele.Value.(*entry).key
		delete(lruCache.cache, key)
		lruCache.currCapacity--
	}
}
