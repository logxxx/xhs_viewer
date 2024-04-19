package main

import (
	"context"
	"fmt"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/modules/heartbeat"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/modules/utils"
	goCache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

func NewNode() *Node {
	node := &Node{agentID2AgentMap: map[string]*Agent{}, cache: goCache.New(30*time.Minute, 1*time.Hour)}
	go node.StatForever()
	return node
}

// GetAgent GetAgent
func (node *Node) GetAgent(agentID string) *Agent {
	node.agentID2AgentMapLock.Lock()
	defer node.agentID2AgentMapLock.Unlock()
	if node.agentID2AgentMap[agentID] != nil {
		return node.agentID2AgentMap[agentID]
	}
	node.agentID2AgentMap[agentID] = NewAgent(node, agentID)

	return node.agentID2AgentMap[agentID]
}

func (node *Node) StatForever() {

	for {
		time.Sleep(10 * time.Second)
		node.agentID2AgentMapLock.Lock()
		count := 0
		for agentID, a := range node.agentID2AgentMap {
			count++
			daemonAddr := "-"
			if a.daemonConn != nil {
				daemonAddr = a.daemonConn.RemoteAddr().String()
			}
			log.Debugf("[NODE_STAT %v/%v]agentID:%v daemonAddr:%v dialID:%v counter[accept:%v dial:%v serve:%v]", count, len(node.agentID2AgentMap), agentID, daemonAddr, a.DialID, a.acceptConnCount, a.callDialApiCount, a.callDirectorApiCount)
		}
		node.agentID2AgentMapLock.Unlock()
	}

}

func (node *Node) StartAccept() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", tcpPort))
	if err != nil {
		panic(err)
	}

	logger := utils.Log(context.Background(), "Node.StartAccept").WithField("tcpPort", tcpPort)

	round := 0
	for {
		logger = logger.WithField("remote_addr", "")
		round++
		logger = logger.WithField("round", round)
		logger.Debugf("Node Accepting...%v", round)

		conn, err := listener.Accept()
		if err != nil {
			logger.Errorf("Accept err:%v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		logger = logger.WithField("remote_addr", conn.RemoteAddr().String())
		logger.Debugf("a conn comein")

		//heartbeat.NewHb().SetData("NODE_SAY_HELLO").Write(conn)

		authData, err := heartbeat.GetAuth(conn)
		if err != nil {
			logger.Errorf("GetAuth err:%v", err)
			continue
		}
		logger.Debugf("get authData:%+v", authData)

		agent := node.GetAgent(authData.DeviceID)

		agent.DealConn(conn, authData.Payload == "I_AM_DAEMON_CONN")

	}

}
