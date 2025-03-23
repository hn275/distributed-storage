package algo

import (
	"container/heap"
	"errors"
)

type LeastConnection struct {
	priorityQueue
}

// LeastConnection implements LBAlgo
func (lc *LeastConnection) Initialize() {
	heap.Init(lc)
}

// LeastConnection implements LBAlgo
func (lc *LeastConnection) NodeJoin(node QueueNode) {
	heap.Push(lc, node)
}

// LeastConnection implements LBAlgo
func (lc *LeastConnection) GetNode() (QueueNode, error) {
	if lc.Len() == 0 {
		return nil, errors.New("queue empty")
	}
	node := heap.Pop(lc).(QueueNode)
	return node, nil
}

// LeastConnection implements LBAlgo
func (lc *LeastConnection) PutNode(node QueueNode) {
	lc.NodeJoin(node)
}
