package algo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeastConnectionQueueNodeCmp(t *testing.T) {
	left := LCNode{nil, 0}
	right := LCNode{nil, 0}

	assert.False(t, left.less(&right))
	assert.False(t, right.less(&left))

	right.connCtr = 1
	assert.True(t, left.less(&right))
}
