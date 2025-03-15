package algo

import (
	"container/heap"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testPQ struct{ float64 }

func (tpq *testPQ) less(other queueNodeCmp) bool {
	otherNode := other.(*testPQ)
	return tpq.float64 < otherNode.float64
}

func TestPriorityQueueSortInterface(t *testing.T) {
	pq := make(priorityQueue, 7)

	pq[0] = queueNode{&testPQ{1.0}}
	pq[1] = queueNode{&testPQ{1.1}}
	pq[2] = queueNode{&testPQ{1.2}}
	pq[3] = queueNode{&testPQ{1.3}}
	pq[4] = queueNode{&testPQ{1.4}}
	pq[5] = queueNode{&testPQ{1.4}}
	pq[6] = queueNode{&testPQ{0.4}}

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
	assert.Equal(t, 1.3, pq[0].node.(*testPQ).float64)
	assert.Equal(t, 1.0, pq[3].node.(*testPQ).float64)
}

func TestPriorityQueueHeapInterface(t *testing.T) {
	const N = 1024
	pq := make(priorityQueue, N)
	expectedValues := make(sort.Float64Slice, N)

	for i := 0; i < N; i++ {
		v := rand.Float64()
		pq[i] = queueNode{&testPQ{v}}
		expectedValues[i] = v
	}

	sort.Sort(expectedValues)

	// test heap.Init
	heap.Init(&pq)

	// test pop in order after init
	for i := 0; i < N; i++ {
		v := heap.Pop(&pq).(queueNode)
		assert.Equal(t, expectedValues[i], v.node.(*testPQ).float64)
		assert.Equal(t, N-i-1, pq.Len())
	}

	// test push and pop in order
	for i := 0; i < N; i++ {
		heap.Push(&pq, queueNode{&testPQ{expectedValues[N-i-1]}})
		assert.Equal(t, i+1, pq.Len())
	}

	for i := 0; i < N; i++ {
		v := heap.Pop(&pq).(queueNode)
		assert.Equal(t, expectedValues[i], v.node.(*testPQ).float64)
		assert.Equal(t, N-i-1, pq.Len())
	}
}
