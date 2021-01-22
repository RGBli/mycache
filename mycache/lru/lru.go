package lru

import (
	"container/list"
)

// lru 类型的 cache
type Cache struct {
	maxBytes     int64
	currentBytes int64
	// 键是字符串，值是指向 list 中元素的指针
	cache map[string]*list.Element
	// 标准库中的双向链表
	list *list.List
}

// 为了使 value 类型更加多样，只要实现了 Value 接口的类型都可以作为 value 的类型
type Value interface {
	Len() int
}

// key-value 对
type entry struct {
	key   string
	value Value
}

// 用于创建一个 Cache 实例
func NewCache(maxBytes int64) *Cache {
	if maxBytes > 0 {
		return &Cache{
			maxBytes: maxBytes,
			cache:    make(map[string]*list.Element),
			list:     list.New(),
		}
	}
	panic("maxBytes cannot be 0")
}

// 用于添加一个 key-value 对到 Cache 中
func (lruCache *Cache) Put(key string, value Value) {
	if ele, ok := lruCache.cache[key]; ok {
		lruCache.list.MoveToFront(ele)
		// 类型断言
		kv := ele.Value.(*entry)
		kv.value = value
		lruCache.currentBytes += int64(value.Len()) - int64(kv.value.Len())
	} else {
		ele := lruCache.list.PushFront(&entry{key: key, value: value})
		lruCache.cache[key] = ele
		lruCache.currentBytes += int64(len(key)) + int64(value.Len())
	}
	for lruCache.currentBytes > lruCache.maxBytes {
		lruCache.RemoveOldest()
	}
}

// 用于从 Cache 中获取指定 key 的 value
func (lruCache *Cache) Get(key string) (value Value, ok bool) {
	if lruCache.IsExists(key) {
		return lruCache.cache[key].Value.(*entry).value, true
	}
	return
}

// 从 Cache 中删除指定 key 的 key-value 对
func (lruCache *Cache) Delete(key string) {
	if lruCache.IsExists(key) {
		ele := lruCache.cache[key]
		value := ele.Value.(*entry).value
		lruCache.list.Remove(ele)
		delete(lruCache.cache, key)
		lruCache.currentBytes -= int64(len(key)) + int64(value.Len())
	}
}

// 判断 Cache 中是否存在某一 key
func (lruCache *Cache) IsExists(key string) bool {
	if _, ok := lruCache.cache[key]; ok {
		return true
	}
	return false
}

// RemoveOldest 用于淘汰 Cache 中的元素
func (lruCache *Cache) RemoveOldest() {
	ele := lruCache.list.Back()
	if ele != nil {
		lruCache.list.Remove(ele)
		key := ele.Value.(*entry).key
		value := ele.Value.(*entry).value
		lruCache.currentBytes -= int64(len(key)) + int64(value.Len())
		delete(lruCache.cache, key)
	}
}
