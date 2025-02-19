package main

import (
	"net"

	"github.com/hn275/distributed-storage/internal/network"
)

type DataNode struct {
	lbSoc net.Conn
}

func nodeJoin(lbAddr string) (*DataNode, error) {
	laddr, err := net.ResolveTCPAddr(network.ProtoTcp4, ":0") // randomize the port
	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveTCPAddr(network.ProtoTcp4, lbAddr)
	if err != nil {
		return nil, err
	}

	// dial and ping LB, notifying node type
	lbSoc, err := net.DialTCP(network.ProtoTcp4, laddr, raddr)
	if err != nil {
		return nil, err
	}

	if _, err := lbSoc.Write([]byte{network.DataNodeJoin}); err != nil {
		return nil, err
	}

	dataNode := &DataNode{
		lbSoc,
	}

	return dataNode, nil
}
