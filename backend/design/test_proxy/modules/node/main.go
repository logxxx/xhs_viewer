package main

import (
	goCache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

var (
	tcpPort     = "5566"
	servePort   = "5565"
	routePrefix = "drive-reverse-proxy-test"
)

type Node struct {
	cache                *goCache.Cache
	agentID2AgentMap     map[string]*Agent
	agentID2AgentMapLock sync.Mutex
}

func main() {

	log.SetLevel(log.DebugLevel)

	log.Debugf("NODE start.")

	node := &Node{agentID2AgentMap: map[string]*Agent{}, cache: goCache.New(30*time.Minute, 1*time.Hour)}

	go node.StartAccept()

	node.StartServe()

}

func write(w http.ResponseWriter, code int, input string) {
	w.Write([]byte(input))
	w.WriteHeader(code)
}
