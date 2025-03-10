package algo

import (
	"container/heap"
	"errors"
	"math"
	"net"
)

type LeastResponseTime struct {
	priorityQueue
}

// LeastResponseTime implements LBAlgo
func (lrt *LeastResponseTime) Initialize() {
	// nop
}

// LeastResponseTime implements LBAlgo
func (lrt *LeastResponseTime) NodeJoin(conn net.Conn) error {
	heap.Push(lrt, queueNode{conn, 0.0})
	return nil
}

// LeastResponseTime implements LBAlgo
func (lrt *LeastResponseTime) GetNode() (net.Conn, error) {
	if lrt.Len() == 0 {
		return nil, errors.New("no node to dispatch")
	}

	node := lrt.priorityQueue[0].Conn

	if lrt.priorityQueue[0].weight == 0 {
		lrt.priorityQueue[0].weight = math.MaxFloat64
		heap.Init(lrt)
	}

	return node, nil
}
