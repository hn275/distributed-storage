package network

import (
	"encoding/binary"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddrByteConv(t *testing.T) {
	addr := net.TCPAddr{
		IP:   []byte{192, 168, 1, 1},
		Port: 0xff00,
	}

	var buf [6]byte
	assert.Nil(t, AddrToBytes(&addr, buf[:]))
	assert.Equal(t, addr.IP[0], buf[0])
	assert.Equal(t, addr.IP[1], buf[1])
	assert.Equal(t, addr.IP[2], buf[2])
	assert.Equal(t, addr.IP[3], buf[3])
	assert.Equal(t, uint16(addr.Port), binary.LittleEndian.Uint16(buf[4:6]))
}
