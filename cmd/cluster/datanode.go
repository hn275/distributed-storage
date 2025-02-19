package main

import (
	"net"
)

type DataNode struct {
	lbSoc net.Conn
}

func makeDataNode(addr string) (*DataNode, error) {
	laddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8000")
	if err != nil {
		return nil, err
	}

	// dial LB node
	lbSoc, err := net.DialTCP("tcp4", laddr, raddr)
	if err != nil {
		return nil, err
	}

	// ping LB node
	if _, err := lbSoc.Write([]byte("ping")); err != nil {
		return nil, err
	}

	dataNode := &DataNode{
		lbSoc,
	}

	return dataNode, nil
}
