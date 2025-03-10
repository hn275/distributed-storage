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
	lrt[i], lrt[j] = lrt[j], lrt[i]
}

// LeastResponseTime implements heap.Interface
// since nodes aren't joining/leaving, no need to implement
func (lrt LeastResponseTime) Push(x any) {
	panic("not implemented")
}

// LeastResponseTime implements heap.Interface
// since nodes aren't joining/leaving, no need to implement
func (lrt LeastResponseTime) Pop() any {
	panic("not implemented")
}
