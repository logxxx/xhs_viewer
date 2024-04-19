package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/modules/heartbeat"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/modules/utils"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
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

			logger := utils.Log(nil, "reverseProxy.Director")

			if req.URL.Scheme == "" {
				req.URL.Scheme = "http"
			}

			parseResult, err := ParsePath(req)
			if err != nil {
				logger.Debugf("ParsePath err:%v", err)
				return
			}
			logger.Debugf("parseResult:%+v", parseResult)

			if parseResult.DeviceID == "" {
				logger.Debugf("ParsePath empty devic_id")
				return
			}

			agent := node.GetAgent(parseResult.DeviceID)

			req.URL.Host = agent.AgentID //透传给dial用

			logger.Debugf("reverseProxy.Director reqUrl:%v", req.URL.String())
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

	logger := utils.Log(ctx, "Node.dialContext").WithField("net", network).WithField("addr", addr)

	logger.Debugf("dialContext waiting for conn")
	agentID, _, _ := net.SplitHostPort(addr)
	if agentID == "" {
		return nil, errors.New("empty agentID")
	}
	logger = logger.WithField("agent_id", agentID)

	agent := node.GetAgent(agentID)

	retryTimes := 0
	for {
		conn, err := agent.waitForNewConn()
		if conn != nil {
			logger.WithField("remote_addr", conn.RemoteAddr().String()).Debugf("dialContext waitForNewConn succ.")
			return conn, nil
		}
		retryTimes++
		logger = logger.WithField("retry_time", retryTimes)
		if retryTimes >= 3 {
			logger.Debugf("waitForNewConn TIMEOUT")
			return nil, err
		}

	}

}

func (agent *Agent) waitForNewConn() (net.Conn, error) {

	logger := utils.Log(nil, "waitForNewConn")

	select {
	case newConn := <-agent.readyForWorkConnChan:
		agent.acceptConnCount++
		logger.WithField("remote_addr", newConn.RemoteAddr().String()).Debugf("recv chan directly")
		return newConn, nil
	default:
	}

	dialID := atomic.AddInt64(&agent.DialID, 1)
	log.Debugf("dialContext send need more conn sig...")
	agent.needMoreConnChan <- dialID
	log.Debugf("dialContext send need more conn sig succ.waiting conn chan...")
	select {
	case conn := <-agent.readyForWorkConnChan:
		agent.acceptConnCount++
		logger.WithField("remote_addr", conn.RemoteAddr().String()).Debugf("recv chan")
		return conn, nil
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
