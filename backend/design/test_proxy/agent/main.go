package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/heartbeat"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strconv"
	"time"
)

var (
	//nodeAddr = "58.221.37.228:5566"
	nodeAddr = "127.0.0.1:5566"
)

type Listener struct {
	needConnChan chan int64
	connChan     chan net.Conn
	daemonConn   net.Conn
}

func main() {

	log.SetLevel(log.DebugLevel)

	log.Debugf("AGENT start.")

	listener := &Listener{connChan: make(chan net.Conn), needConnChan: make(chan int64)}

	go listener.runDaemonConn()

	go listener.StartDial()

	g := gin.Default()
	g.GET("/ping/:id", func(c *gin.Context) {
		log.Debugf("client call ping")
		c.String(200, fmt.Sprintf("pong %v", time.Now()))
	})

	srv := &http.Server{Handler: g}

	err := srv.Serve(listener)
	if err != nil {
		panic(err)
	}

}

func (l *Listener) runDaemonConn() {
	round := 0
	for {
		round++
		log.Debugf("runDaemonConn start dial. round=%v", round)
		conn, err := net.Dial("tcp", nodeAddr)
		if err != nil {
			log.Errorf("Dial err:%v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("dial succ. round=%v", round)

		h := heartbeat.NewHandler(conn)

		modeResp, err := h.Request("I_AM_DAEMON_CONN", true)
		if err != nil {
			log.Errorf("h.Request err:%v", err)
			continue
		}
		log.Debugf("send msg I_AM_DAEMON_CONN to node, resp:%v", modeResp)

		l.startDaemonConn(conn)

	}

}

func (l *Listener) StartDial() {
	round := 0

	for {
		<-l.needConnChan
		round++
		log.Debugf("dial conn start. round=%v nodeAddr=%v", round, nodeAddr)
		conn, err := net.Dial("tcp", nodeAddr)
		if err != nil {
			log.Errorf("Dial err:%v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("dial succ. round=%v", round)

		err = heartbeat.NewHb().SetData("GENERAL_CONN").Write(conn)
		if err != nil {
			log.Errorf("Write err:%v", err)
		}
		l.connChan <- conn

	}
}

func (l *Listener) startDaemonConn(conn net.Conn) {
	l.daemonConn = conn
	defer func() {
		if l.daemonConn != nil {
			l.daemonConn.Close()
			l.daemonConn = nil
		}
	}()

	for {
		hb, err := heartbeat.Read(conn)
		if err != nil {
			return
		}
		log.Debugf("startDeamonConn recv hb:%v", hb.Data())

		if hb.Data() == "KEEP_ALIVE" {
			log.Debugf("recv KEEP_ALIVE msg")
			continue
		}

		log.Debugf("sending need more chan...")
		l.needConnChan <- 1
		log.Debugf("sending need more chan succ!")
	}

}

func (l *Listener) Accept() (net.Conn, error) {
	conn, ok := <-l.connChan
	if !ok {
		log.Debugf("agent Accept conn failed!")
	}
	log.Debugf("agent Accept conn succ.")
	return conn, nil
}

func (l *Listener) Close() error {
	close(l.connChan)
	return nil
}

func (l *Listener) Addr() net.Addr {
	h, p, _ := net.SplitHostPort(nodeAddr)
	addr := &net.TCPAddr{IP: net.ParseIP(h)}
	addr.Port, _ = strconv.Atoi(p)
	return addr
}
