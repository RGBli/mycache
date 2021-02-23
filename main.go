package main

import (
	"flag"
	"io/ioutil"
	"log"
	"mycache"
	"net/http"
	"strconv"
	"strings"
)

// 声明配置参数
var (
	port     string
	peers    []string
	dbNames  []string
	maxBytes int64
)

const localPrefix = "127.0.0.1:"

// 初始化参数
func initPara(path string) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic("Invalid cfg file path")
	}
	s := string(bytes)
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lineSplit := resolveLine(line)
		switch lineSplit[0] {
		case "port":
			port = lineSplit[1]
		case "peers":
			peers = lineSplit[1:]
		case "dbNames":
			dbNames = lineSplit[1:]
		case "maxBytes":
			v := lineSplit[1]
			maxBytes, err = strconv.ParseInt(v[:len(v)-2], 10, 64)
			if err != nil {
				panic("There are format errors in cfg file maxBytes line")
			}
			if strings.HasSuffix(v, "kb") {
				maxBytes *= 1 << 10
			} else if strings.HasSuffix(v, "mb") {
				maxBytes *= 1 << 20
			} else if strings.HasSuffix(v, "gb") {
				maxBytes *= 1 << 30
			}
		}
	}
}

func resolveLine(line string) []string {
	parts := strings.Split(line, " ")
	if len(parts) == 1 {
		panic("There are format errors in cfg file")
	}
	return parts
}

// 启动 cache server
func startCacheServer(port string) {
	server := mycache.NewServer(localPrefix+port, peers)
	server.SetPeers(peers)
	for i := 0; i < len(dbNames); i++ {
		server.NewDatabase(dbNames[i], maxBytes)
	}
	log.Println("MyCache is running at", localPrefix+port)
	log.Fatal(http.ListenAndServe(localPrefix+port, server))
}

func main() {
	var cfgFilePath string
	flag.StringVar(&cfgFilePath, "cfg", "config.cfg", "Configuration file path")
	flag.Parse()
	initPara(cfgFilePath)
	startCacheServer(port)
}
