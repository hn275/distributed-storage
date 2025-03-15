package algo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRTQueueNodeCmp(t *testing.T) {
	left := LRTNode{nil, 0, 0.0}
	right := LRTNode{nil, 0, 0.1}
	assert.True(t, left.less(&right))

	right.avgRT = 0.0
	assert.False(t, left.less(&right))
	assert.False(t, right.less(&left))

	right.requests = 1
	assert.True(t, left.less(&right))
}
