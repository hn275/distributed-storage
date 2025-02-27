package loadbalance

import (
	"log/slog"
	"net"
)

type chanOpCode int

const (
	chanOpCode_nodeJoin = chanOpCode(iota)
	chanOpCode_nodeDispatch
	chanOpCode_response
)

type chanSignal struct {
	opCode chanOpCode
	// nullable
	conn net.Conn
	err  error
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
			lb.serviceNodeJoin(&signal)

		case chanOpCode_nodeDispatch:
			lb.serviceNodeDispatch()

		default:
			slog.Error("unhandled op code", "code", signal.opCode)

		}
	}
}

func (lb *LoadBalancer) serviceNodeDispatch() {
	// query for free node's address
	nodeConn, err := lb.engine.GetNode()

	lb.connChan <- chanSignal{
		opCode: chanOpCode_response,
		conn:   nodeConn,
		err:    err,
	}
}

func (lb *LoadBalancer) serviceNodeJoin(signal *chanSignal) {
	if signal.conn == nil {
		panic("invalid signal, required signal.conn and signal.connChan")
	}

	lb.connChan <- chanSignal{
		opCode: chanOpCode_response,
		conn:   nil,
		err:    lb.engine.NodeJoin(signal.conn),
	}
}
