package loadbalance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleAlgo(t *testing.T) {
	algo := NewSimpleAlgo()
	assert.NotNil(t, algo)
}
