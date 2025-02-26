package loadbalance

import (
	"log/slog"
	"net"
)

type LBAlgo interface {
	// adding a new node to the cluster
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
)

type chanSignal struct {
	opCode int
	// nullable
	connChan chan net.Conn
	// nullable
	conn net.Conn

	err error
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

func (lb *LoadBalancer) NodeJoin(node net.Conn) error {
	nodeJoinSignal := chanSignal{
		chanOpCode_nodeJoin,
		nil,
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
		opCode:   chanOpCode_error,
		connChan: nil,
		conn:     nil,
		err:      lb.engine.NodeJoin(signal.conn),
	}
}
