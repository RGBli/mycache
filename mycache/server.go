package mycache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mycache/consistenthash"
	"net/http"
	"net/url"
	"strings"
)

// 服务器
type Server struct {
	// 包含 IP 和端口，例如 127.0.0.1:8001
	host      string
	databases map[string]*database
	peers     []string
	hashMap   *consistenthash.HashMap
}

/*数据库
* 每个数据库都有一个数据库号，以及对应的 cache*/
type database struct {
	// 数据库名
	dbName    string
	mainCache *cache
}

// 创建一个服务器
func NewServer(host string, peers []string) *Server {
	hashMap := consistenthash.NewMap(3, nil)
	hashMap.AddNodes(peers...)
	return &Server{
		host:      host,
		databases: make(map[string]*database),
		peers:     peers,
		hashMap:   hashMap,
	}
}

const HTTP_PREFIX string = "http://"

// 给服务器创建一个数据库
func (server *Server) NewDatabase(dbName string, maxBytes int64) {
	server.databases[dbName] = &database{
		dbName:    dbName,
		mainCache: &cache{maxBytes: maxBytes},
	}
}

// 设置 peers
func (server *Server) SetPeers(peers []string) {
	server.peers = peers
}

// 添加 peers
func (server *Server) AddPeers(peers []string) {
	for _, val := range peers {
		server.peers = append(server.peers, val)
	}
}

// 获取 key 应该存在的 server 的 host
func (server *Server) GetPeerHost(key string) string {
	return server.hashMap.Get(key)
}

// 从指定的数据库中获取值
func (server *Server) Get(key, dbName string) (ByteView, error) {
	if key == "" || dbName == "" {
		return ByteView{}, fmt.Errorf("Key and database name are required")
	}
	if server.databases[dbName] == nil {
		return ByteView{}, fmt.Errorf("Database %s dose not exist", dbName)
	}
	targetHost := server.GetPeerHost(key)
	if targetHost == server.host {
		value, ok := server.databases[dbName].mainCache.get(key)
		if ok {
			server.Log("Get %s from local host", key)
			return value, nil
		} else {
			server.Log("Key %s does not exist", key)
			return ByteView{}, nil
		}
	}
	return server.getFromPeer(key, dbName)
}

// 从 peer 中获取
func (server *Server) getFromPeer(key, dbName string) (ByteView, error) {
	targetHost := server.GetPeerHost(key)
	url := fmt.Sprintf(
		"%v%v/%v/%v",
		HTTP_PREFIX,
		targetHost,
		url.QueryEscape(dbName),
		url.QueryEscape(key),
	)
	server.Log("Get %s from peer host %s", key, targetHost)
	response, err := http.Get(url)
	if err != nil {
		return ByteView{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return ByteView{}, fmt.Errorf("Server returned: %v", response.Status)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ByteView{}, fmt.Errorf("Reading response body: %v", err)
	}
	return ByteView{bytes: body}, nil
}

// 向指定的数据库添加 k-v 对
func (server *Server) Put(key string, value ByteView, dbName string) error {
	if key == "" || dbName == "" {
		return fmt.Errorf("Key and database name are required")
	}
	if server.databases[dbName] == nil {
		return fmt.Errorf("Database %s dose not exist", dbName)
	}
	targetHost := server.GetPeerHost(key)
	if targetHost == server.host {
		server.Log("Put %s to local host", key)
		server.databases[dbName].mainCache.put(key, value)
		return nil
	}
	return server.putToPeer(key, value, dbName)
}

// 向 peer 中添加
func (server *Server) putToPeer(key string, value ByteView, dbName string) error {
	targetHost := server.GetPeerHost(key)
	url := fmt.Sprintf(
		"%v%v/%v",
		HTTP_PREFIX,
		targetHost,
		url.QueryEscape(dbName),
	)
	body, err := json.Marshal(map[string]string{key: value.ToString()})
	server.Log("Put %s to peer host %s", key, targetHost)
	request, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	_, err = http.DefaultClient.Do(request)
	return err
}

// 从指定的数据库中删除一个 k-v 对
func (server *Server) Delete(key, dbName string) error {
	if key == "" || dbName == "" {
		return fmt.Errorf("Key and database name are required")
	}
	if server.databases[dbName] == nil {
		return fmt.Errorf("Database %s dose not exist", dbName)
	}
	targetHost := server.GetPeerHost(key)
	if targetHost == server.host {
		server.Log("Delete %s from local host", key)
		server.databases[dbName].mainCache.delete(key)
		return nil
	}
	return server.deleteFromPeer(key, dbName)
}

// 从 peer 中删除一个 k-v 对
func (server *Server) deleteFromPeer(key, dbName string) error {
	targetHost := server.GetPeerHost(key)
	url := fmt.Sprintf(
		"%v%v/%v/%v",
		HTTP_PREFIX,
		targetHost,
		url.QueryEscape(dbName),
		url.QueryEscape(key),
	)
	server.Log("Delete %s from peer host %s", key, targetHost)
	request, _ := http.NewRequest("DELETE", url, nil)
	_, err := http.DefaultClient.Do(request)
	return err
}

// 给服务器添加 Log 方法
func (server *Server) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", server.host, fmt.Sprintf(format, v...))
}

// HTTP 服务器的逻辑
func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.Log("%s %s", r.Method, r.URL.Path)
	data := ""
	if r.Method == "GET" {
		parts := strings.SplitN(r.URL.Path[1:], "/", 2)
		if len(parts) != 2 {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		dbName := parts[0]
		key := parts[1]
		value, err := server.Get(key, dbName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data = value.ToString()
	} else if r.Method == "POST" {
		dbName := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		entry := make(map[string]string)
		json.Unmarshal(body, &entry)
		var key string
		var value string
		for k, v := range entry {
			key = k
			value = v
		}
		err = server.Put(key, ByteView{bytes: []byte(value)}, dbName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data = fmt.Sprintf("Put %s succeed", key)
	} else if r.Method == "DELETE" {
		parts := strings.SplitN(r.URL.Path[1:], "/", 2)
		if len(parts) != 2 {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		dbName := parts[0]
		key := parts[1]
		err := server.Delete(key, dbName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data = "Delete Succeed!"
	}
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.Write([]byte(data))
}
