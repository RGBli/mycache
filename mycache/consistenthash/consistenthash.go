package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	// 哈希函数
	hash Hash
	// 每个真实节点对应的虚拟节点的数量
	replicas int
	// 哈希环
	keys []int
	// 虚拟节点和真实节点的映射表，键是虚拟节点的哈希值，值是真实节点的名称
	hashMap map[int]string
}

func NewMap(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if fn == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加真实节点
func (m *Map) AddKeys(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hashValue := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hashValue)
			m.hashMap[hashValue] = key
		}
	}
	sort.Ints(m.keys)
}

// 获取 key 对应的真实节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hashValue := int(m.hash([]byte(key)))
	// 如果没找到 index，则会返回 len(m.keys)
	index := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hashValue
	})
	return m.hashMap[m.keys[index%len(m.keys)]]
}
