package main

import (
	"fmt"
	"github.com/logxxx/xhs_viewer/backend/design/test_proxy/modules/heartbeat"
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
		for agentID, a := range node.agentID2AgentMap {
			log.Debugf("[NODE_STAT]agentID:%v dialID:%v hasDaemon:%v counter[accept:%v dial:%v serve:%v]", agentID, a.DialID, a.daemonConn != nil, a.acceptConnCount, a.callDialApiCount, a.callDirectorApiCount)
		}
		node.agentID2AgentMapLock.Unlock()
	}

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
			log.Errorf("Accept err:%v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("a conn comein:%v", conn.RemoteAddr().String())

		//heartbeat.NewHb().SetData("NODE_SAY_HELLO").Write(conn)

		authData, err := heartbeat.GetAuth(conn)
		if err != nil {
			log.Errorf("GetAuth err:%v", err)
			continue
		}

		//TODO: check auth

		agent := node.GetAgent(authData.DeviceID)

		agent.DealConn(conn, authData.Payload == "I_AM_DAEMON_CONN")

	}

}
