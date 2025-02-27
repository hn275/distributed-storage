package main

import (
	"log/slog"
	"net"

	"github.com/hn275/distributed-storage/internal/network"
)

const LBNodeAddr string = "127.0.0.1:8000"

func main() {
	lbConn, err := net.Dial(network.ProtoTcp4, LBNodeAddr)
	if err != nil {
		panic(err)
	}

	if _, err := lbConn.Write([]byte{network.UserNodeJoin}); err != nil {
		panic(err)
	}

	slog.Info("connected to LB.", "remote_addr", lbConn.RemoteAddr())

	var buf [0xff]byte
	n, err := lbConn.Read(buf[:])
	if err != nil {
		panic(err)
	}

	if n != 6 {
		panic("protocol violation, expecting 6 bytes only.")
	}

	dataNodeAddr, err := network.BytesToAddr(buf[:n])
	if err != nil {
		panic(err)
	}

	dataConn, err := net.DialTCP(network.ProtoTcp4, nil, dataNodeAddr.(*net.TCPAddr))
	if err != nil {
		panic(err)
	}

	slog.Info("data node connected.", "addr", dataConn.RemoteAddr(), "protocol", dataConn.RemoteAddr().Network())
}
