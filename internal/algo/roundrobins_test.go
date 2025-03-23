package algo

import (
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// testStruct implements net.Conn
type testStruct struct {
	id    float64
	index int
}

// SetIndex implements QueueNode.
func (t *testStruct) SetIndex(i int) {
	t.index = i
}

// testStruct implements QueueNode
// this is a nop
func (t *testStruct) Less(QueueNode) bool {
	return false
}

func TestRoundRobinInitialize(t *testing.T) {
	rr := RoundRobin{}

	// tests for init
	rr.Initialize()
	assert.Equal(t, rr.index, 0)
	assert.Equal(t, len(rr.queue), 0)
}

func TestRoundRobinNodeJoin(t *testing.T) {
	rr := RoundRobin{}

	// tests for node join
	rr.NodeJoin(&testStruct{})
	assert.Equal(t, len(rr.queue), 1)
	assert.Equal(t, rr.index, 0)
}

func TestRoundRobinGetNode(t *testing.T) {
	rr := RoundRobin{}

	_, _ = rr.GetNode()

	// node join
	rr.NodeJoin(&testStruct{})
	rr.NodeJoin(&testStruct{})
	assert.Equal(t, len(rr.queue), 2)

	// get node
	node, err := rr.GetNode()
	assert.Nil(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, len(rr.queue), 2)
	assert.Equal(t, rr.index, 1)

	node, err = rr.GetNode()
	assert.Nil(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, len(rr.queue), 2)
	assert.Equal(t, rr.index, 0)
}

func (testt *testStruct) Read(b []byte) (n int, err error) {
	return 0, nil
}
func (testt *testStruct) Write(b []byte) (n int, err error) {
	return 0, nil
}
func (testt *testStruct) Close() error {
	return nil
}
func (testt *testStruct) LocalAddr() net.Addr {
	return net.TCPAddrFromAddrPort(netip.AddrPort{})
}
func (testt *testStruct) RemoteAddr() net.Addr {
	return net.TCPAddrFromAddrPort(netip.AddrPort{})
}
func (testt *testStruct) SetDeadline(t time.Time) error {
	return nil
}
func (testt *testStruct) SetReadDeadline(t time.Time) error {
	return nil
}
func (testt *testStruct) SetWriteDeadline(t time.Time) error {
	return nil
}
