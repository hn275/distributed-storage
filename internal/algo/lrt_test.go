package algo

import (
	"container/heap"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeastResponseTimeSortInterface(t *testing.T) {
	lrt := make(LeastResponseTime, 7)

	lrt[0] = &LRTNode{&testStruct{}, 1.0}
	lrt[1] = &LRTNode{&testStruct{}, 1.1}
	lrt[2] = &LRTNode{&testStruct{}, 1.2}
	lrt[3] = &LRTNode{&testStruct{}, 1.3}
	lrt[4] = &LRTNode{&testStruct{}, 1.4}
	lrt[5] = &LRTNode{&testStruct{}, 1.4}
	lrt[6] = &LRTNode{&testStruct{}, 0.4}

	// Len()
	assert.Equal(t, len(lrt), lrt.Len())

	// Less(i, j int) bool
	assert.True(t, lrt.Less(0, 1))
	assert.True(t, lrt.Less(0, 2))
	assert.True(t, lrt.Less(0, 3))
	assert.False(t, lrt.Less(1, 0))
	assert.False(t, lrt.Less(2, 0))
	assert.False(t, lrt.Less(3, 0))

	// testing equality
	assert.False(t, lrt.Less(4, 5))
	assert.False(t, lrt.Less(5, 4))

	// testing transitivity

	// Swap(i, j int)
	lrt.Swap(0, 3)
	assert.Equal(t, 1.3, lrt[0].avgResponseTime)
	assert.Equal(t, 1.0, lrt[3].avgResponseTime)

	// heapify
	heap.Init(lrt)
	log.Println("heap.Init:")
	for i, v := range lrt {
		fmt.Println(i, v.avgResponseTime)
	}
}
