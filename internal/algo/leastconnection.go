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

// LeastConnection implements LBAlgo
func (lc *LeastConnection) Fix(i int) error {
	if i >= lc.priorityQueue.Len() {
		return errors.New("index i out of bound.")
	}

	heap.Fix(&lc.priorityQueue, i)
	return nil
}

// LeastConnection implements LBAlgo
func (lc *LeastConnection) Queue() []QueueNode {
	return lc.priorityQueue
}
