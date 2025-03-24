package network

import (
	"encoding/binary"
	"fmt"
	"net"
)

// message types
const (
	DataNodeJoin = iota
	UserNodeJoin
	PortForwarding
	HealthCheck
	ShutdownSig

	ProtoTcp4       = "tcp4"
	RandomLocalPort = "127.0.0.1:0"
)

var (
	BinaryEndianess = binary.LittleEndian
)

func AddrToBytes(addr net.Addr, buf []byte) error {
	if len(buf) < 6 {
		return fmt.Errorf("insufficient buf size")
	}

	addr_, ok := addr.(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("returned type is not net.TCPAddr: %v", addr)
	}

	ip := addr_.IP.To4()
	if ip == nil {
		return fmt.Errorf("returned type is not net.TCPAddr: %v", addr)
	}

	copy(buf[:4], ip)

	binary.LittleEndian.PutUint16(buf[4:6], uint16(addr_.Port))

	return nil
}

// parsing the first 6 bytes for `buf` only
func BytesToAddr(buf []byte) (net.Addr, error) {
	if len(buf) < 6 {
		return nil, fmt.Errorf("insufficient buf length")
	}

	addr := &net.TCPAddr{
		IP:   buf[:4],
		Port: int(binary.LittleEndian.Uint16(buf[4:6])),
	}

	return addr, nil
}
