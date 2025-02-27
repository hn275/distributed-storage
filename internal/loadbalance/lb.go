package loadbalance

import (
	"io"
	"net"

	"github.com/hn275/distributed-storage/internal/loadbalance/algo"
	"github.com/hn275/distributed-storage/internal/network"
)

type LoadBalancer struct {
	net.Listener
	engine   algo.LBAlgo
	connChan chan chanSignal
}

func New(protocol, addr string, algo algo.LBAlgo) (*LoadBalancer, error) {
	soc, err := net.Listen(protocol, addr)

	if err != nil {
		return nil, err
	}

	loadBalancer := &LoadBalancer{
		soc,
		algo,
		make(chan chanSignal),
	}

	go loadBalancer.queryServer()

	return loadBalancer, nil
}

func (lb *LoadBalancer) UserHandler(user net.Conn) error {
	defer user.Close()

	// request for a data node
	lb.connChan <- chanSignal{chanOpCode_nodeDispatch, nil, nil}

	sig := <-lb.connChan
	if sig.err != nil {
		return sig.err
	}

	// port fowarding buffer
	// 2 bytes for the port uint16 port number
	var buf [2]byte

	// ping data node
	buf[0] = network.UserNodeJoin

	_, err := sig.conn.Write(buf[0:1])
	if err != nil {
		return err
	}

	// forward the 2 byte port number to user
	_, err = io.CopyN(user, sig.conn, 6)

	return err
}

func (lb *LoadBalancer) NodeJoin(node net.Conn) error {
	nodeJoinSignal := chanSignal{
		chanOpCode_nodeJoin,
		node,
		nil,
	}

	lb.connChan <- nodeJoinSignal

	return (<-lb.connChan).err
}

func (lb *LoadBalancer) NodeDispatch() (net.Conn, error) {
	nodeJoinSignal := chanSignal{
		chanOpCode_nodeDispatch,
		nil,
		nil,
	}

	lb.connChan <- nodeJoinSignal
	response := <-lb.connChan

	return response.conn, response.err
}
