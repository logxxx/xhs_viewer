package main

import (
	"context"
	"fmt"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/modules/heartbeat"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/modules/utils"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type Agent struct {
	AgentID    string
	ParentNode *Node
	DialID     int64

	readyForWorkConnChan chan net.Conn
	daemonConn           net.Conn
	needMoreConnChan     chan int64

	acceptConnCount      int
	callDialApiCount     int
	callDirectorApiCount int
}

// NewAgent NewAgent
func NewAgent(node *Node, id string) *Agent {
	resp := &Agent{
		ParentNode:           node,
		AgentID:              id,
		DialID:               0,
		readyForWorkConnChan: make(chan net.Conn),
		needMoreConnChan:     make(chan int64),
	}

	go resp.reportForever()
	go resp.watchForNeedMoreChan()
	return resp
}

func (a *Agent) watchForNeedMoreChan() {
	logger := utils.Log(context.Background(), "watchForNeedMoreChan").WithField("agent_id", a.AgentID)
	round := 0
	for {
		round++
		logger = logger.WithField("round", round)
		select {
		case dialID := <-a.needMoreConnChan:
			logger = logger.WithField("dialID", dialID)
			ok := a.needMoreConnByDaemon(dialID)
			if ok {
				logger.Debugf("needMoreConnByDaemon succ!")
				continue
			}
			logger.Debugf("needMoreConnByDaemon failed.")
		}
	}
}

func (a *Agent) DealConn(conn net.Conn, isDaemonConn bool) {

	logger := utils.Log(context.Background(), "DealConn").WithField("remote_addr", conn.RemoteAddr().String())

	if isDaemonConn {
		a.startDaemonConn(conn)
		return
	}

	timeoutTick := time.NewTicker(5 * time.Minute)

	keepAliveTick := time.NewTicker(30 * time.Second)

	select {
	case <-keepAliveTick.C:
		err := keepAlive(conn)
		if err != nil {
			logger.Errorf("keepAlive failed:%v", err)
			return
		}
		logger.Debugf("keepAlive succ")
		keepAliveTick.Reset(30 * time.Second)
	case <-timeoutTick.C:
		logger.Debugf("recv a chan but TOO LONG no use")
	case a.readyForWorkConnChan <- conn:
		logger.Debugf("DealConn push conn to readyForWorkConnChan succ")
	}

}

func keepAlive(conn net.Conn) error {
	return heartbeat.NewHb().SetData("KEEP_ALIVE").Write(conn)
}

func (a *Agent) needMoreConnByMqtt(dialID int64) (err error) {

	_, ok := a.ParentNode.cache.Get(fmt.Sprintf("call_mqtt_interval_%v", a.AgentID))
	if !ok {
		return nil
	}

	//TODO intarval config
	a.ParentNode.cache.Set(fmt.Sprintf("call_mqtt_interval_%v", a.AgentID), time.Now().Unix(), time.Minute*15)

	//TODO: CALL MQTT

	return
}

func (a *Agent) needMoreConnByDaemon(dialID int64) (ok bool) {

	logger := utils.Log(context.Background(), "needMoreConnByDaemon").WithField("dialID", dialID)

	if a.daemonConn == nil {
		logger.Debugf("failed: a.daemonConn == nil")
		return
	}

	if !isConnHealthy(a.daemonConn) {
		logger.Debugf("failed: isConnHealthy is false")
		return
	}

	req := fmt.Sprintf("NEED_MORE:%v", dialID)
	log.Debugf("startDaemonConn recv need more chan sig. so ask agent:%v", req)
	_, err := heartbeat.NewHandler(a.daemonConn).Request(req, false)
	if err != nil {
		log.Debugf("Request err:%v", err)
		return
	}
	log.Debugf("send need more succ.")

	ok = true

	return

}

func (a *Agent) startDaemonConn(conn net.Conn) {

	err := heartbeat.NewHb().SetData("OK").Write(conn)
	if err != nil {
		log.Errorf("startDaemonConn Write err:%v", err)
		return
	}

	//keep alive
	go func() {
		for {
			time.Sleep(10 * time.Second)

			log.Debugf("startDaemonConn send keep alive msg...")
			err = keepAlive(conn)
			if err == nil {
				log.Debugf("startDaemonConn keep alive succ!")
				a.daemonConn = conn
				continue
			}

			log.Errorf("startDaemonConn send KEEP_ALIVE failed.err:%v", err)
			if a.daemonConn == conn {
				a.daemonConn = nil
			}
			return
		}
	}()

}

func (a *Agent) reportForever() {
	for {
		time.Sleep(30 * time.Second)
		log.Debugf("direct:%v dial:%v conn:%v daemonConnExist:%v", a.callDirectorApiCount, a.callDialApiCount, a.acceptConnCount, a.daemonConn != nil)
	}
}
