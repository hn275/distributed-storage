package internal_test

import (
	"testing"

	"github.com/hn275/distributed-storage/internal"
	"github.com/stretchr/testify/assert"
)

func TestCalcMovingAvg(t *testing.T) {
	expected := []struct {
		currentAvg   float64
		currentValue float64
		expected     float64
	}{
		{0.0, 1.0, 1.0},
		{1.0, 4.0, 2.5},
		{2.5, 4, 3.0},
	}

	for i, v := range expected {
		result := internal.CalcMovingAvg(uint64(i), v.currentAvg, v.currentValue)
		assert.Equal(t, v.expected, result)
	}
}
