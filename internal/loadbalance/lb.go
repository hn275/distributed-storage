package loadbalance

import (
	"log/slog"
	"net"

	"github.com/hn275/distributed-storage/internal/network"
)

type LBAlgo interface {
	Initialize()
	NodeJoin(net.Conn) error
	GetNode() (net.Conn, error)
}

type LoadBalancer struct {
	net.Listener
	engine LBAlgo

	// cursed
	connChan chan chanSignal
}

const (
	chanOpCode_nodeJoin = iota
	chanOpCode_error
	chanOpCode_nodeRequest
)

type chanSignal struct {
	opCode int
	// nullable
	conn net.Conn
	err  error
}

func NewBalancer(protocol, addr string, algo LBAlgo) (*LoadBalancer, error) {
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
	// query for free node's address
	node, err := lb.engine.GetNode()
	if err != nil {
		return err
	}

	var buf [0xff]byte
	buf[0] = network.UserNodeJoin

	_, err = node.Write(buf[0:1])
	if err != nil {
		return err
	}

	// forward address
	n, err := node.Read(buf[:])
	if err != nil {
		return err
	}

	slog.Info("node's listening server.", "msg", string(buf[:n]))

	_, err = user.Write(buf[:n])
	if err != nil {
		return err
	}

	return nil
}

func (lb *LoadBalancer) NodeJoin(node net.Conn) error {
	nodeJoinSignal := chanSignal{
		chanOpCode_nodeJoin,
		node,
		nil,
	}

	lb.connChan <- nodeJoinSignal

	response := <-lb.connChan
	if response.opCode != chanOpCode_error {
		panic("expected signal for error response")
	}

	return response.err
}

func (lb *LoadBalancer) NodeDispatch() (net.Conn, error) {
	nodeJoinSignal := chanSignal{
		chanOpCode_nodeRequest,
		nil,
		nil,
	}

	lb.connChan <- nodeJoinSignal
	response := <-lb.connChan

	return response.conn, response.err
}

func (lb *LoadBalancer) queryServer() {
	defer func() {
		slog.Info("closing query server.")
		close(lb.connChan)
	}()

	for {
		slog.Info("query server waiting for request")
		signal := <-lb.connChan

		switch signal.opCode {

		case chanOpCode_nodeJoin:
			lb.handleNodeJoin(&signal)

		default:
			slog.Error("unhandled op code", "code", signal.opCode)

		}
	}
}

func (lb *LoadBalancer) handleNodeJoin(signal *chanSignal) {
	if signal.conn == nil {
		panic("invalid signal, required signal.conn and signal.connChan")
	}

	lb.connChan <- chanSignal{
		opCode: chanOpCode_error,
		conn:   nil,
		err:    lb.engine.NodeJoin(signal.conn),
	}
}
