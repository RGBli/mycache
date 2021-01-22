package main

import (
	"log"
	"mycache"
	"net/http"
)

func main() {
	mycache.NewDatabase(1, 2<<10)
	socket := "localhost:9999"
	peers := mycache.NewHTTPPool(socket)
	log.Println("mycache is running at", socket)
	log.Fatal(http.ListenAndServe(socket, peers))
}
