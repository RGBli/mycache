package lru

import (
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestCache_Get(t *testing.T) {
	lruCache := NewCache(int64(0))
	lruCache.Put("lbw", String("22"))
	value, ok := lruCache.Get("lbw")
	if !ok || string(value.(String)) != "22" {
		t.Fatalf(string(value.(String)))
	}
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
