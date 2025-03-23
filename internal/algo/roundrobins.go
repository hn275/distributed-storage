package algo

import (
	"fmt"
)

// RoundRobin implements LBAlgo.
type RoundRobin struct {
	queue []QueueNode
	index int
}

func (rr *RoundRobin) Initialize() {
	rr.queue = make([]QueueNode, 0)
	rr.index = 0
}

func (rr *RoundRobin) NodeJoin(node QueueNode) {
	rr.queue = append(rr.queue, node)
}

func (rr *RoundRobin) GetNode() (QueueNode, error) {
	if len(rr.queue) == 0 {
		return nil, fmt.Errorf("no node can be scheduled.")
	}

	node := rr.queue[rr.index]
	rr.index = (rr.index + 1) % len(rr.queue)

	return node, nil
}

func (rr *RoundRobin) PutNode(node QueueNode) {
	// nop
}

func (rr *RoundRobin) Fix(int) error {
	// nop
	return nil
}
