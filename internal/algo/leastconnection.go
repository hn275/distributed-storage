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
func (left *LCNode) less(other queueNodeCmp) bool {
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
func (lrt *LeastConnection) NodeJoin(conn net.Conn) error {
	qNode := queueNode{
		node: &LCNode{conn, 0},
	}
	heap.Push(lrt, qNode)
	return nil
}

// LeastConnection implements LBAlgo
func (lrt *LeastConnection) GetNode() (net.Conn, error) {
	if lrt.Len() == 0 {
		return nil, errors.New("no node to dispatch")
	}

	node := heap.Pop(lrt).(*LCNode)
	node.connCtr += 1

	heap.Push(lrt, node)
	return node, nil
}
