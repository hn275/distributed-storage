package algo

import (
	"container/heap"
	"math/rand"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testPQ struct{ float64 }

// LocalAddr implements QueueNode.
func (t *testPQ) LocalAddr() net.Addr {
	panic("unimplemented")
}

// Read implements QueueNode.
func (t *testPQ) Read(b []byte) (n int, err error) {
	panic("unimplemented")
}

// RemoteAddr implements QueueNode.
func (t *testPQ) RemoteAddr() net.Addr {
	panic("unimplemented")
}

// SetDeadline implements QueueNode.
func (*testPQ) SetDeadline(t time.Time) error {
	panic("unimplemented")
}

// SetReadDeadline implements QueueNode.
func (*testPQ) SetReadDeadline(t time.Time) error {
	panic("unimplemented")
}

// SetWriteDeadline implements QueueNode.
func (*testPQ) SetWriteDeadline(t time.Time) error {
	panic("unimplemented")
}

// Write implements QueueNode.
func (t *testPQ) Write(b []byte) (n int, err error) {
	panic("unimplemented")
}

// testPQ implements net.Conn
func (t *testPQ) Close() error {
	return nil
}

func (tpq *testPQ) Less(other QueueNode) bool {
	otherNode := other.(*testPQ)
	return tpq.float64 < otherNode.float64
}

func TestPriorityQueueSortInterface(t *testing.T) {
	pq := make(priorityQueue, 7)

	pq[0] = &testPQ{1.0}
	pq[1] = &testPQ{1.1}
	pq[2] = &testPQ{1.2}
	pq[3] = &testPQ{1.3}
	pq[4] = &testPQ{1.4}
	pq[5] = &testPQ{1.4}
	pq[6] = &testPQ{0.4}

	// Len()
	assert.Equal(t, len(pq), pq.Len())

	// Less(i, j int) bool
	assert.True(t, pq.Less(0, 1))
	assert.True(t, pq.Less(0, 2))
	assert.True(t, pq.Less(0, 3))
	assert.False(t, pq.Less(1, 0))
	assert.False(t, pq.Less(2, 0))
	assert.False(t, pq.Less(3, 0))

	// testing equality
	assert.False(t, pq.Less(4, 5))
	assert.False(t, pq.Less(5, 4))

	// Swap(i, j int)
	pq.Swap(0, 3)
	assert.Equal(t, 1.3, pq[0].(*testPQ).float64)
	assert.Equal(t, 1.0, pq[3].(*testPQ).float64)
}

func TestPriorityQueueHeapInterface(t *testing.T) {
	const N = 1024
	pq := make(priorityQueue, N)
	expectedValues := make(sort.Float64Slice, N)

	for i := 0; i < N; i++ {
		v := rand.Float64()
		pq[i] = &testPQ{v}
		expectedValues[i] = v
	}

	sort.Sort(expectedValues)

	// test heap.Init
	heap.Init(&pq)

	// test pop in order after init
	for i := 0; i < N; i++ {
		v := heap.Pop(&pq).(QueueNode)
		assert.Equal(t, expectedValues[i], v.(*testPQ).float64)
		assert.Equal(t, N-i-1, pq.Len())
	}

	// test push and pop in order
	for i := 0; i < N; i++ {
		heap.Push(&pq, &testPQ{expectedValues[N-i-1]})
		assert.Equal(t, i+1, pq.Len())
	}

	for i := 0; i < N; i++ {
		v := heap.Pop(&pq).(QueueNode)
		assert.Equal(t, expectedValues[i], v.(*testPQ).float64)
		assert.Equal(t, N-i-1, pq.Len())
	}
}
