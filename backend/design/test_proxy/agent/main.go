package main

import (
	"flag"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/heartbeat"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	//nodeAddr = "58.221.37.228:5566"
	//nodeAddr = "127.0.0.1:5566"
	nodeAddr = flag.String("node_addr", "", "")
)

type Listener struct {
	needConnChan chan int64
	connChan     chan net.Conn
	daemonConn   net.Conn
	authData     *heartbeat.AuthData
}

func NewListener(authData *heartbeat.AuthData) *Listener {
	listener := &Listener{connChan: make(chan net.Conn), needConnChan: make(chan int64), authData: authData}
	return listener
}

func main() {

	flag.Parse()

	log.SetLevel(log.DebugLevel)

	log.Debugf("AGENT start.")

	if *nodeAddr == "" {
		panic("empty node addr")
	}

	listener := NewListener(&heartbeat.AuthData{
		UserID:        "",
		DeviceID:      "",
		Authorization: "",
		Payload:       "",
	})

	go listener.runDaemonConn()

	go listener.StartDial()

	//g := gin.Default()
	//g.GET("/ping/:id", func(c *gin.Context) {
	//	log.Debugf("client call ping")
	//	c.String(200, fmt.Sprintf("pong %v", time.Now()))
	//})
	//

	srv := &http.Server{Handler: &PortProxy{}}

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
		conn, err := net.Dial("tcp", *nodeAddr)
		if err != nil {
			log.Errorf("Dial err:%v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("dial succ. round=%v", round)

		h := heartbeat.NewHandler(conn)

		authData := l.authData.DeepCopy()
		authData.Payload = "I_AM_DAEMON_CONN"
		err = h.Auth(authData)
		if err != nil {
			log.Errorf("h.Auth err:%v", err)
			continue
		}

		l.startDaemonConn(conn)

	}

}

func (l *Listener) StartDial() {
	round := 0

	for {
		dialID := <-l.needConnChan
		round++
		log.Debugf("dial conn start. dialID=%v round=%v nodeAddr=%v", dialID, round, *nodeAddr)
		conn, err := net.Dial("tcp", *nodeAddr)
		if err != nil {
			log.Errorf("Dial err:%v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("dial succ. round=%v", round)

		h := heartbeat.NewHandler(conn)
		authData := l.authData.DeepCopy()
		authData.Payload = "I_AM_GENERAL_CONN"
		err = h.Auth(authData)
		if err != nil {
			log.Errorf("h.Auth err:%v", err)
			continue
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

		//	req := fmt.Sprintf("NEED_MORE:%v", dialID)

		if strings.HasPrefix(hb.Data(), "NEED_MORE:") {
			dialID, _ := strconv.ParseInt(strings.TrimPrefix(hb.Data(), "NEED_MORE:"), 10, 64)
			log.Debugf("dial_id=%v sending need more chan...", dialID)
			l.needConnChan <- dialID
			log.Debugf("dial_id=%v sending need more chan succ!", dialID)
			continue
		}

		log.Debugf("recv invalid data:%v", hb.Data())

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
	h, p, _ := net.SplitHostPort(*nodeAddr)
	addr := &net.TCPAddr{IP: net.ParseIP(h)}
	addr.Port, _ = strconv.Atoi(p)
	return addr
}
