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

func (httpPool *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, httpPool.baseUrl) {
		panic("HTTP Server serving an unexpected path: " + r.URL.Path)
	}
	log.Printf("%s %s", r.Method, r.URL.Path)

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
			http.Error(w, "key not found", http.StatusNotFound)
			return
		}

		value, err2 := db.Get(key)
		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(value + "\n"))

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

		db := NewDatabase(int8(dbNumber), 100)
		for key, value := range entries {
			db.Put(key, value)
		}

	} else if r.Method == "DELETE" {
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
			http.Error(w, "key not found", http.StatusNotFound)
			return
		}

	} else if r.Method == "HEAD" {

	}
}
