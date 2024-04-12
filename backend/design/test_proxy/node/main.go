package main

import (
	"context"
	"fmt"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/heartbeat"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

var (
	tcpPort     = "5566"
	servePort   = "5565"
	routePrefix = "drive-reverse-proxy-test"
)

type Node struct {
	acceptConnCount      int
	callDialApiCount     int
	callDirectorApiCount int
	connChan             chan net.Conn
	daemonConn           net.Conn
	needMoreConnChan     chan int64
	dialID               int64
}

func main() {

	log.SetLevel(log.DebugLevel)

	log.Debugf("NODE start.")

	node := &Node{connChan: make(chan net.Conn), needMoreConnChan: make(chan int64)}

	go node.StartAccept()

	go node.reportForever()

	node.StartServe()

}

func (node *Node) StartServe() {

	server := http.Server{Handler: node.reverseProxy()}
	for {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%v", servePort))
		if err != nil {
			panic(err)
		}

		server.Serve(listener)
	}
}

func (node *Node) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	node.reverseProxy().ServeHTTP(rw, req)
}

func (node *Node) reportForever() {
	for {
		time.Sleep(5 * time.Second)
		log.Debugf("direct:%v dial:%v conn:%v daemonConnExist:%v", node.callDirectorApiCount, node.callDialApiCount, node.acceptConnCount, node.daemonConn != nil)
	}
}

func (node *Node) reverseProxy() *httputil.ReverseProxy {
	resp := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			node.callDirectorApiCount++
			//TODO: parse path

			pathElems := strings.Split(req.URL.Path, "/")
			if len(pathElems) <= 3 {
				log.Debugf("err: len(pathElems) <= 3: %v", pathElems)
				return
			}

			if pathElems[1] != routePrefix {
				log.Debugf("pathElems[1] != routePrefix elem:%v prefix:%v", pathElems[1], routePrefix)
				return
			}

			log.Debugf("before parse path, req.URL.Path:%v", req.URL.Path)
			req.URL.Path = "/" + filepath.Join(pathElems[2:]...)
			log.Debugf("after parse path, req.URL.Path:%v", req.URL.Path)

			if req.URL.Scheme == "" {
				req.URL.Scheme = "http"
			}
			req.URL.Host = "127.0.0.1"

			log.Debugf("reverseProxy.Director reqUrl:%v", req.URL.String())
		},
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          1024,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
			MaxIdleConnsPerHost:   300,
			DialContext:           node.dialContext,
		},
	}
	return resp
}

func (node *Node) dialContext(ctx context.Context, network, addr string) (conn net.Conn, err error) {
	log.Debugf("dialContext waiting for conn")
	node.callDialApiCount++
	dialID := atomic.AddInt64(&node.dialID, 1)
	select {
	case newConn := <-node.connChan:
		conn = newConn
		log.Debugf("dialContext get conn directly")
		return
	default:
	}
	log.Debugf("dialContext send need more conn sig...")
	node.needMoreConnChan <- dialID
	log.Debugf("dialContext send need more conn sig succ.waiting conn chan...")
	conn = <-node.connChan
	//TODO: before return, check if conn is healthy (by KEEP_ALIVE or other msg)
	log.Debugf("dialContext conn start work")
	node.acceptConnCount++
	return
}

func (node *Node) StartAccept() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", tcpPort))
	if err != nil {
		panic(err)
	}

	round := 0
	for {
		round++
		log.Debugf("Node Accepting...%v", round)
		conn, err := listener.Accept()
		if err != nil {
			log.Debugf("Accept err:%v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("a conn comein:%v", conn.RemoteAddr().String())

		//heartbeat.NewHb().SetData("NODE_SAY_HELLO").Write(conn)

		firstHb, err := heartbeat.Read(conn)
		if err != nil {
			log.Debugf("Read err:%v", err)
			continue
		}

		log.Printf("recv firstHb:%v", firstHb.Data())

		if firstHb.Data() == "I_AM_DAEMON_CONN" {
			node.startDaemonConn(conn)
		} else {
			select {
			case <-time.After(3 * time.Minute):
				log.Debugf("recv a chan but TOO LONG no use")
			case node.connChan <- conn:
			}
		}

	}

}

func (node *Node) startDaemonConn(conn net.Conn) {

	err := heartbeat.NewHb().SetData("OK").Write(conn)
	if err != nil {
		log.Errorf("startDaemonConn Write err:%v", err)
	}

	h := heartbeat.NewHandler(conn)

	//keep alive
	go func() {
		for {
			time.Sleep(10 * time.Second)

			log.Debugf("startDaemonConn send keep alive msg...")
			_, err := h.Request("KEEP_ALIVE", false)
			if err == nil {
				log.Debugf("startDaemonConn keep alive succ!")
				node.daemonConn = conn
				continue
			}

			log.Errorf("startDaemonConn send KEEP_ALIVE failed.err:%v", err)
			if node.daemonConn == conn {
				node.daemonConn = nil
			}
			return
		}
	}()

	go func() {
		for {
			dialID := <-node.needMoreConnChan
			req := fmt.Sprintf("NEED_MORE:%v", dialID)
			log.Debugf("startDaemonConn recv need more chan sig. so ask agent:%v", req)
			_, err := h.Request(req, false)
			if err != nil {
				log.Debugf("startDaemonConn Request err:%v", err)
				return
			}
			log.Debugf("startDaemonConn send need more succ.")
		}
	}()

}

func write(w http.ResponseWriter, code int, input string) {
	w.Write([]byte(input))
	w.WriteHeader(code)
}
