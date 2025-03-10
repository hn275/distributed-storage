package algo

import (
	"math"
	"net"
)

type LBAlgo interface {
	Initialize()
	NodeJoin(net.Conn) error
	GetNode() (net.Conn, error)
}

type queueNode struct {
	net.Conn
	weight float64
}

type priorityQueue []queueNode

func (pq priorityQueue) Len() int {
	return len(pq)
}

// LeastResponseTime implements sort.Interface. This math assume NaN to be the
// smallest value, see https://pkg.go.dev/sort#Float64Slice.Less
func (pq priorityQueue) Less(i, j int) bool {
	return (pq[i].weight < pq[j].weight) ||
		(math.IsNaN(pq[i].weight) && !math.IsNaN(pq[j].weight))
}

// LeastResponseTime implements sort.Interface
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

// LeastResponseTime implements heap.Interface
func (pq *priorityQueue) Push(x any) {
	node, ok := x.(queueNode)
	if !ok {
		panic("invalid interface, expected `queueNode`")
	}
	*pq = append(*pq, node)
}

// LeastResponseTime implements heap.Interface
// since nodes aren't joining/leaving, no need to implement
func (pq *priorityQueue) Pop() any {
	n := len(*pq) - 1
	node := (*pq)[n]
	*pq = (*pq)[:n]
	return node
}
