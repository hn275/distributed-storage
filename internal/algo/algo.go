package algo

import "net"

type LBAlgo interface {
	Initialize()
	NodeJoin(net.Conn) error
	GetNode() (net.Conn, error)
}
