package algo

import (
	"math"
	"net"
)

type LRTNode struct {
	net.Conn
	avgResponseTime float64
}

type LeastResponseTime []*LRTNode

// LeastResponseTime implements sort.Interface
func (lrt LeastResponseTime) Len() int {
	return len(lrt)
}

// LeastResponseTime implements sort.Interface. This math assume NaN to be the
// smallest value, see https://pkg.go.dev/sort#Float64Slice.Less
func (lrt LeastResponseTime) Less(i, j int) bool {
	return (lrt[i].avgResponseTime < lrt[j].avgResponseTime) ||
		(math.IsNaN(lrt[i].avgResponseTime) && !math.IsNaN(lrt[j].avgResponseTime))
}

// LeastResponseTime implements sort.Interface
func (lrt LeastResponseTime) Swap(i, j int) {
	tmp := lrt[i]
	lrt[i] = lrt[j]
	lrt[j] = tmp
}

// LeastResponseTime implements heap.Interface
func (lrt LeastResponseTime) Push(node *LRTNode) {}

// LeastResponseTime implements heap.Interface
func (lrt LeastResponseTime) Pop() *LRTNode {
	return nil
}
