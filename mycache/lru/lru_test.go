package lru

import (
	"fmt"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestCache_Get(t *testing.T) {
	lruCache := NewCache(1 << 10)
	lruCache.Put("lbw", String("22"))
	fmt.Println(lruCache.cache["lbw"].Value.(*entry).value)
}

func TestCache_Delete(t *testing.T) {
	lruCache := NewCache(int64(0))
	lruCache.Put("lbw", String("22"))
	lruCache.Put("sg", String("3"))
	lruCache.Delete("lbw")
	value, ok := lruCache.Get("lbw")
	if ok {
		t.Fatalf(string(value.(String)))
	}
}

func TestCache_RemoveOldest(t *testing.T) {
	lruCache := NewCache(2)
	lruCache.Put("sg", String("3"))
	lruCache.Put("lbw", String("22"))
	if _, ok := lruCache.Get("sg"); ok {
		t.Fatal("Error")
	}
}
