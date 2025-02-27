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

	dataNodeAddr := string(buf[:n])
	dataConn, err := net.Dial(network.ProtoTcp4, dataNodeAddr)
	if err != nil {
		panic(err)
	}

	slog.Info("data node connected.", "addr", dataConn.RemoteAddr(), "protocol", dataConn.RemoteAddr().Network())
}
