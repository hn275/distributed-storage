package algo

import (
	"container/heap"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testPQ struct{ float64 }

func (tpq *testPQ) less(other queueNodeCmp) bool {
	otherNode := other.(*testPQ)
	return tpq.float64 < otherNode.float64
}

func TestPriorityQueue(t *testing.T) {
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

	// heapify
	heap.Init(&pq)
	assert.Equal(t, 0.4, pq[0].node.(*testPQ).float64)

	// mod the first element
	pq[4] = queueNode{&testPQ{0.0}}

	heap.Init(&pq)
	assert.Equal(t, float64(0), pq[0].node.(*testPQ).float64)

	// min value popping
	expectedMinSequence := [7]float64{
		0.0,
		0.4,
		1.0,
		1.1,
		1.2,
		1.3,
		1.4,
	}

	for i, expected := range expectedMinSequence {
		node := heap.Pop(&pq).(queueNode)
		assert.Equal(t, expected, node.node.(*testPQ).float64)
		assert.Equal(t, 7-i-1, pq.Len())
		assert.Equal(t, 7-i-1, len(pq))
	}
}
