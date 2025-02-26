package loadbalance

import (
	"fmt"
	"net"

	"github.com/eapache/queue"
)

type RoundRobin struct {
	queue *queue.Queue
}

// RoundRobin implements LBAlgo.

func (rr *RoundRobin) Initialize() {
	rr.queue = queue.New()
}

func (rr *RoundRobin) NodeJoin(node net.Conn) error {
	rr.queue.Add(node)
	return nil
}
func (rr *RoundRobin) GetNode() (net.Conn, error) {
	if rr.queue.Length() == 0 {
		return nil, fmt.Errorf("no node can be scheduled.")
	}

	node, ok := rr.queue.Remove().(net.Conn)
	if !ok {
		panic("queue should only contain net.Conn elements.")
	}

	return node, nil
}
