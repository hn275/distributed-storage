package loadbalance

import (
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testStruct struct{}

func TestRoundRobinInitialize(t *testing.T) {
	rr := RoundRobin{}

	// tests for init
	rr.Initialize()
	assert.Equal(t, rr.size, 0)
	assert.Equal(t, rr.index, 0)
	assert.Equal(t, len(rr.queue), 0)
}

func TestRoundRobinNodeJoin(t *testing.T) {
	rr := RoundRobin{}

	// tests for node join
	err := rr.NodeJoin(&testStruct{})
	assert.Nil(t, err)
	assert.Equal(t, rr.size, 1)
	assert.Equal(t, rr.index, 0)
}

func TestRoundRobinGetNode(t *testing.T) {
	rr := RoundRobin{}

	_, _ = rr.GetNode()

	// node join
	_ = rr.NodeJoin(&testStruct{})
	_ = rr.NodeJoin(&testStruct{})
	assert.Equal(t, rr.size, 2)

	// get node
	node, err := rr.GetNode()
	assert.Nil(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, rr.size, 2)
	assert.Equal(t, rr.index, 1)

	node, err = rr.GetNode()
	assert.Nil(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, rr.size, 2)
	assert.Equal(t, rr.index, 0)
}

// testStruct implements net.Conn
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
