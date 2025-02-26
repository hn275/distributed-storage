package main

import (
	"fmt"
	"net"

	"github.com/hn275/distributed-storage/internal/network"
)

const LBNodeAddr string = "127.0.0.1:8000"

func main() {
	lbConn, err := net.Dial(network.ProtoTcp4, LBNodeAddr)
	if err != nil {
		panic(err)
	}

	pingMsg := [...]byte{network.UserNodeJoin}
	if _, err := lbConn.Write(pingMsg[:]); err != nil {
		panic(err)
	}

	buf := make([]byte, 128)
	n, err := lbConn.Read(buf)
	if err != nil {
		panic(err)
	}

	buf = buf[:n]
	fmt.Println("LB response", string(buf))
}
