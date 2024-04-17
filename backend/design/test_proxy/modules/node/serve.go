package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/modules/heartbeat"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

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

func (node *Node) reverseProxy() *httputil.ReverseProxy {
	resp := &httputil.ReverseProxy{
		Director: func(req *http.Request) {

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

			parseResult := ParsePath(req)

			if parseResult.DeviceID == "" {
				log.Debugf("ParsePath empty devic_id")
				return
			}

			agent := node.GetAgent(parseResult.DeviceID)

			req.URL.Host = agent.AgentID //透传给dial用

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
	agentID, _, _ := net.SplitHostPort(addr)
	if agentID == "" {
		return nil, errors.New("empty agentID")
	}

	agent := node.GetAgent(agentID)

	retryTimes := 0
	for {
		conn, err := agent.waitForNewConn()
		if conn != nil {
			return conn, nil
		}
		retryTimes++
		if retryTimes >= 3 {
			return nil, err
		}

	}

}

func (agent *Agent) waitForNewConn() (net.Conn, error) {
	select {
	case newConn := <-agent.connChan:
		if isConnHealthy(newConn) {
			log.Debugf("dialContext get conn directly")
			agent.acceptConnCount++
			return newConn, nil
		}
	default:
	}

	dialID := atomic.AddInt64(&agent.dialID, 1)
	log.Debugf("dialContext send need more conn sig...")
	agent.needMoreConnChan <- dialID
	log.Debugf("dialContext send need more conn sig succ.waiting conn chan...")
	select {
	case conn := <-agent.connChan:
		if isConnHealthy(conn) {
			log.Debugf("dialContext get conn start work")
			agent.acceptConnCount++
			return conn, nil
		}
	case <-time.After(3 * time.Second):
	}

	return nil, errors.New("no available connection")
}

func isConnHealthy(conn net.Conn) (isHealthy bool) {
	err := heartbeat.NewHb().SetData("PING").Write(conn)
	if err != nil {
		log.Errorf("isConnHealthy Write err:%v", err)
		return
	}
	hb, err := heartbeat.Read(conn)
	if err != nil {
		log.Errorf("isConnHealthy Read err:%v", err)
		return
	}
	if hb.Data() != "OK" {
		log.Errorf("isConnHealthy recv invalid resp:%v", hb.Data())
		return
	}
	return true
}
