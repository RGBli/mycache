package mycache

import (
	"fmt"
	"log"
	"sync"
)

/*数据库
* 每个数据库都有一个数据库号，以及对应的 cache*/
type Database struct {
	number    int8
	mainCache *cache
	peers     PeerPicker
}

var (
	mu        sync.RWMutex
	databases = make(map[int8]*Database)
)

// 创建数据库
func NewDatabase(number int8, maxBytes int64) *Database {
	mu.Lock()
	defer mu.Unlock()
	db := &Database{
		number:    number,
		mainCache: &cache{maxBytes: maxBytes},
	}
	databases[number] = db
	return db
}

// 获取数据库
func GetDatabase(number int8) *Database {
	mu.RLock()
	db := databases[number]
	mu.RUnlock()
	return db
}

// 从指定数据库中获取某一 key 的 value
func (db *Database) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if value, ok := db.mainCache.get(key); ok {
		fmt.Println("[mycache] hit")
		return value, nil
	}
	if db.peers != nil {
		if peer, ok := db.peers.PickPeer(key); ok {
			value, err := db.getFromPeer(peer, key)
			if err == nil {
				return value, nil
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}
	return ByteView{}, nil
}

func (db *Database) RegisterPeers(peers PeerPicker) {
	if db.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	db.peers = peers
}

func (db *Database) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(uint8(db.number), key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{bytes: bytes}
	db.Put(key, value)
	return value, nil
}

// 向指定数据库插入 key-value 对
func (db *Database) Put(key string, value ByteView) {
	if key == "" {
		fmt.Println("key is required")
		return
	}
	db.mainCache.put(key, value)
}

// 从指定数据库删除 key-value 对
func (db *Database) Delete(key string) {
	if key == "" {
		fmt.Println("key is required")
		return
	}
	db.mainCache.delete(key)
}

// 判断指定数据库中是否存在某一 key-value 对
func (db *Database) IsExists(key string) bool {
	if key == "" {
		fmt.Println("key is required")
		return false
	}
	return db.mainCache.isExists(key)
}
