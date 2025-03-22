package algo

import (
	"container/heap"
	"errors"
	"net"
)

// A node for least connection algo
type LCNode struct {
	net.Conn
	// connection counter
	connCtr uint16
}

// LCNode implements queueNodeCmp
func (left *LCNode) Less(other QueueNode) bool {
	right, ok := other.(*LCNode)
	if !ok {
		panic("invalid type, expected '*LCNode'")
	}

	return left.connCtr < right.connCtr
}

type LeastConnection struct {
	priorityQueue
}

// LeastConnection implements LBAlgo
func (lrt *LeastConnection) Initialize() {
	heap.Init(lrt)
}

// LeastConnection implements LBAlgo
func (lrt *LeastConnection) NodeJoin(node QueueNode) {
	heap.Push(lrt, node)
}

// LeastConnection implements LBAlgo
func (lrt *LeastConnection) GetNode() (QueueNode, error) {
	if lrt.Len() == 0 {
		return nil, errors.New("queue empty")
	}

	node := heap.Pop(lrt).(*LCNode)
	node.connCtr += 1

	heap.Push(lrt, node)
	return node, nil
}
