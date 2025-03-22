package algo

import (
	"container/heap"
	"errors"
	"net"
)

// A node for least response time algo
type LRTNode struct {
	net.Conn
	requests uint64
	// average response time
	avgRT float64
}

// LRTNode implements queueNodeCmp
func (left *LRTNode) Less(other QueueNode) bool {
	right, ok := other.(*LRTNode)
	if !ok {
		panic("invalid type, expected '*LRTNode'")
	}

	if left.avgRT == right.avgRT {
		return left.requests < right.requests
	}

	return left.avgRT < right.avgRT
}

type LeastResponseTime struct {
	priorityQueue
}

// LeastResponseTime implements LBAlgo
func (lrt *LeastResponseTime) Initialize() {
	heap.Init(lrt)
}

// LeastResponseTime implements LBAlgo
func (lrt *LeastResponseTime) NodeJoin(node QueueNode) error {
	heap.Push(lrt, node)
	return nil
}

// LeastResponseTime implements LBAlgo
func (lrt *LeastResponseTime) GetNode() (net.Conn, error) {
	if lrt.Len() == 0 {
		return nil, errors.New("no node to dispatch")
	}

	node := heap.Pop(lrt).(*LRTNode)
	node.requests += 1

	heap.Push(lrt, node)
	return node, nil
}
