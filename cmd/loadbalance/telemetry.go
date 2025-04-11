package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	eventUserJoin    = "user-joined"
	eventNodeJoin    = "node-joined"
	eventPortForward = "port-forward"
	eventHealthCheck = "health-check"

	peerUser     = "user"
	peerDataNode = "node"
)

var csvheaders = []string{
	"event-type",
	"peer",
	"node-id",
	"timestamp",
	"duration(ns)",
	"avgRT(ns)",
	"active-requests",
	"queue",
}

type event struct {
	eType     string
	peer      string
	peerID    int32
	timestamp time.Time
	duration  int64
	avgRT     float64
	activeReq uint64
	queue     string
}

// not thread safe at all
func makeQueueString(lb *loadBalancer) string {
	q := lb.engine.Queue()
	s := make([]string, len(q))

	for i, n := range q {
		node := n.(*dataNode)
		s[i] = node.String()
	}

	return strings.Join(s, ", ")
}

func (e *event) Row() []string {
	return []string{
		e.eType,
		e.peer,
		fmt.Sprintf("%d", e.peerID),
		fmt.Sprintf("%d", e.timestamp.UnixNano()),
		fmt.Sprintf("%d", e.duration),
		fmt.Sprintf("%f", e.avgRT),
		fmt.Sprintf("%d", e.activeReq),
		e.queue,
	}
}
