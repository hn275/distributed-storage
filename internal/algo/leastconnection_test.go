package algo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeastConnectionQueueNodeCmp(t *testing.T) {
	left := LCNode{nil, 0}
	right := LCNode{nil, 0}

	assert.False(t, left.Less(&right))
	assert.False(t, right.Less(&left))

	right.connCtr = 1
	assert.True(t, left.Less(&right))
}
