package mycache

import (
	"fmt"
	"sync"
)

type Database struct {
	number    int8
	mainCache *cache
}

var (
	mu        sync.RWMutex
	databases = make(map[int8]*Database)
)

func NewDatabase(number int8, maxCapacity int) *Database {
	mu.Lock()
	defer mu.Unlock()
	db := &Database{
		number:    number,
		mainCache: &cache{maxCapacity: maxCapacity},
	}
	databases[number] = db
	return db
}

func GetDatabase(number int8) *Database {
	mu.RLock()
	db := databases[number]
	mu.RUnlock()
	return db
}

func (db *Database) Get(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key is required")
	}
	if value, ok := db.mainCache.get(key); ok {
		fmt.Println("[mycache] hit")
		return value, nil
	}
	return "", fmt.Errorf("key not exists")
}

func (db *Database) Put(key, value string) {
	if key == "" {
		fmt.Println("key is required")
		return
	}
	db.mainCache.add(key, value)
}
