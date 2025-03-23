package algo

import (
	"container/heap"
	"errors"
)

type LeastResponseTime struct {
	priorityQueue
}

// LeastResponseTime implements LBAlgo
func (lrt *LeastResponseTime) Initialize() {
	heap.Init(lrt)
}

// LeastResponseTime implements LBAlgo
func (lrt *LeastResponseTime) NodeJoin(node QueueNode) {
	heap.Push(lrt, node)
}

// LeastResponseTime implements LBAlgo
func (lrt *LeastResponseTime) GetNode() (QueueNode, error) {
	if lrt.Len() == 0 {
		return nil, errors.New("queue empty")
	}
	node := heap.Pop(lrt).(QueueNode)
	return node, nil
}

// LeastResponseTime implements LBAlgo
func (lrt *LeastResponseTime) PutNode(node QueueNode) {
	lrt.NodeJoin(node)
}
