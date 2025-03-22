package algo

type LBAlgo interface {
	Initialize()
	NodeJoin(QueueNode)
	GetNode() (QueueNode, error)
}

type QueueNode interface {
	less(QueueNode) bool
}

type priorityQueue []QueueNode

// priorityQueue implements sort.Interface
func (pq priorityQueue) Less(i, j int) bool {
	left, right := pq[i], pq[j]
	return left.less(right)
}

// priorityQueue implements sort.Interface
func (pq priorityQueue) Swap(i, j int) {
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
