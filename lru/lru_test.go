package lru

import (
	"testing"
)

func TestPut(t *testing.T) {
	lruCache := NewCache(100)
	lruCache.Put("lbw", "22")
	value, _ := lruCache.Get("lbw")
	if value != "22" {
		t.Fatal("Error")
	}
}

func TestRemoveoldest(t *testing.T) {
	lruCache := NewCache(2)
	lruCache.Put("lbw", "22")
	lruCache.Put("sg", "3")
	if _, ok := lruCache.Get("lbw"); ok {
		t.Fatal("Error")
	}
}
