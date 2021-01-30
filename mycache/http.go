package mycache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mycache/consistenthash"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/mycache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	url         string
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(socket string) *HTTPPool {
	return &HTTPPool{
		url:      socket,
		basePath: defaultBasePath,
	}
}

func (httpPool *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", httpPool.url, fmt.Sprintf(format, v...))
}

func (httpPool *HTTPPool) Set(peers ...string) {
	httpPool.mu.Lock()
	defer httpPool.mu.Unlock()
	httpPool.peers = consistenthash.NewMap(defaultReplicas, nil)
	httpPool.peers.AddKeys(peers...)
	httpPool.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		httpPool.httpGetters[peer] = &httpGetter{baseUrl: peer + httpPool.basePath}
	}
}

// PickPeer picks a peer according to key
func (httpPool *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	httpPool.mu.Lock()
	defer httpPool.mu.Unlock()
	if peer := httpPool.peers.Get(key); peer != "" && peer != httpPool.url {
		httpPool.Log("Pick peer %s", peer)
		return httpPool.httpGetters[peer], true
	}
	return nil, false
}

// 为 HTTPPool 添加 ServeHTTP() 方法，意味着 HTTPPool 实现了 handle 接口
func (httpPool *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, httpPool.basePath) {
		panic("HTTP Server serving an unexpected path: " + r.URL.Path)
	}
	httpPool.Log("%s %s", r.Method, r.URL.Path)
	var data string = "ok"

	if r.Method == "GET" {
		parts := strings.SplitN(r.URL.Path[len(httpPool.basePath):], "/", 2)
		if len(parts) != 2 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		dbNumber, err := strconv.ParseInt(parts[0], 10, 8)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		key := parts[1]
		db := GetDatabase(int8(dbNumber))
		if db == nil {
			http.Error(w, "database not found", http.StatusNotFound)
			return
		}

		value, err := db.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data = value.ToString()

	} else if r.Method == "PUT" {
		dbNumber, err := strconv.Atoi(r.URL.Path[len(httpPool.basePath):])
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		content, _ := ioutil.ReadAll(r.Body)
		var entries map[string]string
		json.Unmarshal(content, &entries)

		db := GetDatabase(int8(dbNumber))
		if db == nil {
			http.Error(w, "database not found", http.StatusNotFound)
			return
		}

		for key, value := range entries {
			db.Put(key, ByteView{[]byte(value)})
		}

	} else if r.Method == "DELETE" {
		parts := strings.SplitN(r.URL.Path[len(httpPool.basePath):], "/", 2)
		if len(parts) != 2 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		dbNumber, err := strconv.ParseInt(parts[0], 10, 8)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		key := parts[1]

		db := GetDatabase(int8(dbNumber))
		if db == nil {
			http.Error(w, "key not found", http.StatusNotFound)
			return
		}

		db.Delete(key)
	}

	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.Write([]byte(data))
}

type httpGetter struct {
	baseUrl string
}

func (h *httpGetter) Get(db uint8, key string) ([]byte, error) {
	url := fmt.Sprintf(
		"%v%v/%v",
		h.baseUrl,
		url.QueryEscape(strconv.Itoa(int(db))),
		url.QueryEscape(key),
	)

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", response.Status)
	}

	bytes, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, err
}

var _ PeerGetter = (*httpGetter)(nil)
