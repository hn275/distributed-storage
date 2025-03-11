package algo

import (
	"container/heap"
	"errors"
	"net"
)

type LRTNode struct {
	net.Conn
	requests uint64
	// average response time
	avgRT float64
}

type LeastResponseTime []*LRTNode

// LeastResponseTime implements sort.Interface. This math assume NaN to be the
// smallest value, see https://pkg.go.dev/sort#Float64Slice.Less
func (lrt LeastResponseTime) Less(i, j int) bool {
	left, right := lrt[i], lrt[j]

	if left.avgRT == right.avgRT {
		return left.requests < right.requests
	}

	return left.avgRT < right.avgRT
}

// LeastResponseTime implements sort.Interface
func (lrt LeastResponseTime) Swap(i, j int) {
	lrt[i], lrt[j] = lrt[j], lrt[i]
}

// priorityQueue implements heap.Interface
func (lrt LeastResponseTime) Len() int {
	return len(lrt)
}

// LeastResponseTime implements heap.Interface
func (lrt *LeastResponseTime) Push(x any) {
	node, ok := x.(*LRTNode)
	if !ok {
		panic("invalid interface, expected `queueNode`")
	}
	heap.Push(lrt, node)
}

// LeastResponseTime implements heap.Interface
// since nodes aren't joining/leaving, no need to implement
func (pq *LeastResponseTime) Pop() any {
	n := len(*pq) - 1
	node := (*pq)[n]
	*pq = (*pq)[:n]
	return node
}

// LeastResponseTime implements LBAlgo
func (lrt *LeastResponseTime) Initialize() {
	heap.Init(lrt)
}

// LeastResponseTime implements LBAlgo
func (lrt *LeastResponseTime) NodeJoin(conn net.Conn) error {
	node := &LRTNode{
		Conn:     conn,
		requests: 0,
		avgRT:    0.0,
	}
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
