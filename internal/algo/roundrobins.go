package algo

import (
	"fmt"
	"net"
)

// RoundRobin implements LBAlgo.
type RoundRobin struct {
	queue []net.Conn
	index int
	size  int
}

func (rr *RoundRobin) Initialize() {
	rr.size = 0
	rr.queue = make([]net.Conn, rr.size)
	rr.index = 0
}

func (rr *RoundRobin) NodeJoin(node net.Conn) error {
	rr.queue = append(rr.queue, node)
	rr.size++
	return nil
}
func (rr *RoundRobin) GetNode() (net.Conn, error) {
	if len(rr.queue) == 0 {
		return nil, fmt.Errorf("no node can be scheduled.")
	}

	node := rr.queue[rr.index]
	rr.index = (rr.index + 1) % rr.size

	return node, nil
}
