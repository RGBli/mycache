package mycache

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const defaultBaseUrl = "/mycache/"

type HTTPPool struct {
	socket  string
	baseUrl string
}

func NewHTTPPool(socket string) *HTTPPool {
	return &HTTPPool{
		socket:  socket,
		baseUrl: defaultBaseUrl,
	}
}

// 为 HTTPPool 添加 ServeHTTP() 方法，意味着 HTTPPool 实现了 handle 接口
func (httpPool *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, httpPool.baseUrl) {
		panic("HTTP Server serving an unexpected path: " + r.URL.Path)
	}
	log.Printf("%s %s", r.Method, r.URL.Path)
	var data string = "ok"

	if r.Method == "GET" {
		parts := strings.SplitN(r.URL.Path[len(httpPool.baseUrl):], "/", 2)
		if len(parts) != 2 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		dbNumber, err1 := strconv.ParseInt(parts[0], 10, 8)
		if err1 != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		key := parts[1]
		db := GetDatabase(int8(dbNumber))
		if db == nil {
			http.Error(w, "database not found", http.StatusNotFound)
			return
		}

		value, err2 := db.Get(key)
		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusInternalServerError)
			return
		}

		data = value.ToString()

	} else if r.Method == "PUT" {
		dbNumber, err := strconv.Atoi(r.URL.Path[len(httpPool.baseUrl):])
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
		parts := strings.SplitN(r.URL.Path[len(httpPool.baseUrl):], "/", 2)
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
