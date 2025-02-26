package loadbalance

import (
	"fmt"
	"net"
)

type SimpleAlgo struct {
	// buffer to hold all the data nodes in the cluster
	// TODO: Emily can define this data structure later, but for now just
	// a map of address -> connection interface
	nodes map[net.Addr]net.Conn
}

// NodeJoin implements LBAlgo.
func (s *SimpleAlgo) NodeJoin(newConn net.Conn) error {
	s.nodes[newConn.RemoteAddr()] = newConn
	fmt.Println(s.nodes)
	return nil
}

// NodeJoin implements LBAlgo.
func (s *SimpleAlgo) GetNode() (net.Conn, error) {
	return nil, nil
}

func NewSimpleAlgo() *SimpleAlgo {
	return &SimpleAlgo{
		nodes: map[net.Addr]net.Conn{},
	}
}
