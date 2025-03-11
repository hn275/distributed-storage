package algo

/*
func TestPriorityQueue(t *testing.T) {
	pq := make(priorityQueue, 7)

	pq[0] = queueNode{nil, 1.0}
	pq[1] = queueNode{nil, 1.1}
	pq[2] = queueNode{nil, 1.2}
	pq[3] = queueNode{nil, 1.3}
	pq[4] = queueNode{nil, 1.4}
	pq[5] = queueNode{nil, 1.4}
	pq[6] = queueNode{nil, 0.4}

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
	assert.Equal(t, 1.3, pq[0].weight)
	assert.Equal(t, 1.0, pq[3].weight)

	// heapify
	heap.Init(&pq)
	assert.Equal(t, 0.4, pq[0].weight)

	// mod the first element
	pq[4] = queueNode{nil, 0.0}

	heap.Init(&pq)
	assert.Equal(t, float64(0), pq[0].weight)

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
		assert.Equal(t, expected, node.weight)
		assert.Equal(t, 7-i-1, pq.Len())
		assert.Equal(t, 7-i-1, len(pq))
	}
}
*/
