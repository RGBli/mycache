package main

import (
	"flag"
	"log"
	"mycache"
	"net/http"
)

func startCacheServer(thisHost string, peerHosts []string, db *mycache.Database) {
	peers := mycache.NewHTTPPool(thisHost)
	peers.Set(peerHosts...)
	db.RegisterPeers(peers)
	log.Println("mycache is running at", thisHost)
	log.Fatal(http.ListenAndServe(thisHost[7:], peers))
}

func startAPIServer(apiHost string, db *mycache.Database) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			value, err := db.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain;charset=utf-8")
			w.Write(value.ByteSlice())

		}))
	log.Println("fontend server is running at", apiHost)
	log.Fatal(http.ListenAndServe(apiHost[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server")
	flag.Parse()

	apiHost := "http://localhost:9999"
	hostMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var hosts []string
	for _, v := range hostMap {
		hosts = append(hosts, v)
	}

	// maxBytes = (1 << 10) B = 1MB
	db := mycache.NewDatabase(1, 1<<10)
	if api {
		go startAPIServer(apiHost, db)
	}
	startCacheServer(hostMap[port], hosts, db)
}
