package main

import (
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/modules/utils"
	goCache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
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

	log.SetFormatter(&utils.MyLogFormatter{})

	log.SetLevel(log.DebugLevel)

	log.Debugf("NODE start.")

	node := NewNode()

	go node.StartAccept()

	node.StartServe()

}

func write(w http.ResponseWriter, code int, input string) {
	w.Write([]byte(input))
	w.WriteHeader(code)
}
