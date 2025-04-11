package algo

import "net"

const (
	AlgoSimpleRoundRobin  = "simple-round-robin"
	AlgoLeastResponseTime = "least-response-time"
	AlgoLeastConnections  = "least-connections"
)

type LBAlgo interface {
	Initialize()
	NodeJoin(QueueNode)
	GetNode() (QueueNode, error)
	PutNode(QueueNode)
	Fix(int) error
	Queue() []QueueNode
}

type QueueNode interface {
	net.Conn
	Less(QueueNode) bool
	SetIndex(i int)
}

type priorityQueue []QueueNode

// priorityQueue implements sort.Interface
func (pq priorityQueue) Less(i, j int) bool {
	left, right := pq[i], pq[j]
	return left.Less(right)
}

// priorityQueue implements sort.Interface
func (pq priorityQueue) Swap(i, j int) {
	// swap the indexes
	pq[i].SetIndex(j)
	pq[j].SetIndex(i)

	// swap
	pq[i], pq[j] = pq[j], pq[i]
}

// priorityQueue implements sort.Interface
func (pq priorityQueue) Len() int {
	return len(pq)
}

// priorityQueue implements heap.Interface
func (pq *priorityQueue) Push(x any) {
	node, ok := x.(QueueNode)
	if !ok {
		panic("invalid interface, expected `queueNode`")
	}

	node.SetIndex(len(*pq))
	*pq = append(*pq, node)
}

// priorityQueue implements heap.Interface
// should returns a QueueNode
func (pq *priorityQueue) Pop() any {
	n := len(*pq) - 1
	node := (*pq)[n]
	*pq = (*pq)[:n]
	return node
}
