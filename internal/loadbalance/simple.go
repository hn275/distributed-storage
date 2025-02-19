package loadbalance

import "net"

type SimpleAlgo struct {
	// buffer to hold all the data nodes in the cluster
	// TODO: Emily can define this data structure later, but for now just
	// a map of address -> connection interface
	nodes map[net.Addr]net.Conn
}

// NodeJoin implements LBAlgo.
func (s *SimpleAlgo) NodeJoin(newConn net.Conn) {
	s.nodes[newConn.RemoteAddr()] = newConn
}

func NewSimpleAlgo() *SimpleAlgo {
	return &SimpleAlgo{
		nodes: map[net.Addr]net.Conn{},
	}
}
