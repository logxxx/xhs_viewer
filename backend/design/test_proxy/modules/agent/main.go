package main

import (
	"context"
	"flag"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/modules/heartbeat"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/modules/utils"
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
	needConnChan         chan int64
	readyForWorkConnChan chan net.Conn
	daemonConn           net.Conn
	authData             *heartbeat.AuthData
}

func NewListener(authData *heartbeat.AuthData) *Listener {
	listener := &Listener{readyForWorkConnChan: make(chan net.Conn), needConnChan: make(chan int64), authData: authData}
	return listener
}

func main() {

	flag.Parse()

	log.SetFormatter(&utils.MyLogFormatter{})

	log.SetLevel(log.DebugLevel)

	log.Debugf("AGENT start.")

	if *nodeAddr == "" {
		panic("empty node addr")
	}

	listener := NewListener(&heartbeat.AuthData{
		UserID:        "test_user1",
		DeviceID:      "device_id#123456",
		Authorization: "auth eyjh123123",
		Payload:       "payload hello world",
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

	logger := utils.Log(context.Background(), "Listener.StartDial")
	for {
		dialID := <-l.needConnChan
		round++
		logger = logger.WithField("dialID", dialID).WithField("round", round)
		logger.Debugf("dial conn start")
		conn, err := net.Dial("tcp", *nodeAddr)
		if err != nil {
			logger.Errorf("Dial err:%v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		logger.Debugf("dial succ")

		h := heartbeat.NewHandler(conn)
		authData := l.authData.DeepCopy()
		authData.Payload = "I_AM_GENERAL_CONN"
		err = h.Auth(authData)
		if err != nil {
			logger.Errorf("h.Auth err:%v", err)
			continue
		}

		logger.Debugf("send to readyForWorkConnChan...")
		l.readyForWorkConnChan <- conn
		logger.Debugf("send to readyForWorkConnChan succ!")
	}
}

func (l *Listener) startDaemonConn(conn net.Conn) {

	logger := utils.Log(context.Background(), "Listener.startDaemonConn").WithField("local_addr", conn.LocalAddr())

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
			logger.Errorf("Read err:%v", err)
			return
		}
		log.Debugf("startDeamonConn recv hb:%v", hb.Data())

		if hb.Data() == "KEEP_ALIVE" {
			logger.Debugf("recv KEEP_ALIVE msg")
			continue
		}

		if hb.Data() == "PING" {
			logger.Debugf("recv PING msg")
			err = heartbeat.NewHb().SetData("OK").Write(conn)
			if err != nil {
				logger.Errorf("Write PING err:%v", err)
				return
			}
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
	conn, ok := <-l.readyForWorkConnChan
	if !ok {
		log.Debugf("agent Accept conn failed!")
	}
	log.Debugf("agent Accept conn succ.")
	return conn, nil
}

func (l *Listener) Close() error {
	close(l.readyForWorkConnChan)
	return nil
}

func (l *Listener) Addr() net.Addr {
	h, p, _ := net.SplitHostPort(*nodeAddr)
	addr := &net.TCPAddr{IP: net.ParseIP(h)}
	addr.Port, _ = strconv.Atoi(p)
	return addr
}
