package algo

import (
	"fmt"
	"net"
)

type LBAlgo interface {
	Initialize()
	NodeJoin(net.Conn) error
	GetNode() (net.Conn, error)
}

func New(algorithm string) (LBAlgo, error) {
	switch algorithm {
	case "rr":
		return &RoundRobin{}, nil

	case "lrt":
		return &LeastResponseTime{}, nil

	case "lc":
		return &LeastConnection{}, nil

	default:
		return nil, fmt.Errorf(
			"invalid algorithm, expected [rr|lrt|lc], got [%s]", algorithm,
		)
	}
}

type queueNodeCmp interface {
	less(queueNodeCmp) bool
}

type queueNode struct {
	node queueNodeCmp
}

type priorityQueue []queueNode

// priorityQueue implements sort.Interface
func (pq priorityQueue) Less(i, j int) bool {
	left, right := pq[i], pq[j]
	return left.node.less(right.node)
}

// priorityQueue implements sort.Interface
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

// priorityQueue implements sort.Interface
func (pq priorityQueue) Len() int {
	return len(pq)
}

// priorityQueue implements heap.Interface
func (pq *priorityQueue) Push(x any) {
	node, ok := x.(queueNode)
	if !ok {
		panic("invalid interface, expected `queueNode`")
	}
	*pq = append(*pq, node)
}

// priorityQueue implements heap.Interface
func (pq *priorityQueue) Pop() any {
	n := len(*pq) - 1
	node := (*pq)[n]
	*pq = (*pq)[:n]
	return node
}
