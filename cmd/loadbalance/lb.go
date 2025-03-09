package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/hn275/distributed-storage/internal/algo"
	"github.com/hn275/distributed-storage/internal/network"
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

type loadBalancer struct {
	net.Listener
	engine   algo.LBAlgo
	connChan chan chanSignal
}

// query server

func (lb *loadBalancer) queryServer() {
	defer func() {
		slog.Info("closing query server.")
		close(lb.connChan)
	}()

	slog.Info("query server waiting for requests")
	for signal := range lb.connChan {
		switch signal.opCode {

		case chanOpCode_nodeJoin:
			lb.nodeJoinQuery(&signal)

		case chanOpCode_nodeDispatch:
			lb.nodeDispatchQuery()

		default:
			lb.connChan <- chanSignal{
				opCode: chanOpCode_response,
				conn:   nil,
				err:    fmt.Errorf("unhandled op code: %d", signal.opCode),
			}

		}
	}
}

func (lb *loadBalancer) nodeDispatchQuery() {
	// query for free node's address
	nodeConn, err := lb.engine.GetNode()

	if err != nil {
		lb.connChan <- chanSignal{
			opCode: chanOpCode_response,
			conn:   nil,
			err:    err,
		}

		return
	}

	if nodeConn == nil {
		lb.connChan <- chanSignal{
			opCode: chanOpCode_response,
			conn:   nil,
			err:    fmt.Errorf("nodeConn nil??"),
		}
		return
	}

	lb.connChan <- chanSignal{
		opCode: chanOpCode_response,
		conn:   nodeConn,
		err:    err,
	}
}

func (lb *loadBalancer) nodeJoinQuery(signal *chanSignal) {
	if signal.conn == nil {
		panic("invalid signal, required signal.conn and signal.connChan")
	}

	lb.connChan <- chanSignal{
		opCode: chanOpCode_response,
		conn:   nil,
		err:    lb.engine.NodeJoin(signal.conn),
	}
}

// server handlers

// listener
func (lbSrv *loadBalancer) listen() {
	var buf [0xff]byte
	for {
		conn, err := lbSrv.Accept()
		if err != nil {
			logger.Error("failed to accept new conn.",
				"peer", conn.RemoteAddr,
				"err", err,
			)
			continue
		}

		if _, err = conn.Read(buf[:]); err != nil {
			// silent continue if peer disconnected
			if !errors.Is(err, io.EOF) {
				logger.Error("failed to read from socket.",
					"remote_addr", conn.RemoteAddr(),
					"err", err,
				)
			}
			continue
		}

		switch buf[0] {
		case network.DataNodeJoin:
			logger.Info("new data node.", "remote_addr", conn.RemoteAddr())
			go handle(conn, lbSrv.nodeJoinHandler)

		case network.UserNodeJoin:
			logger.Info("new user.", "remote_addr", conn.RemoteAddr())
			go handle(conn, lbSrv.userJoinHandler)

		case network.ShutdownSig:
			return

		default:
			logger.Error("unsupported ping message type.", "msgtype", buf[0])
			closeConn(conn)
		}
	}
}

func (lb *loadBalancer) userJoinHandler(user net.Conn) error {
	defer user.Close()

	// request for a data node
	lb.connChan <- chanSignal{chanOpCode_nodeDispatch, nil, nil}

	sig := <-lb.connChan
	if sig.err != nil {
		return sig.err
	}

	if sig.conn == nil {
		return fmt.Errorf("nil node connection.")
	}

	// port fowarding
	_, err := sig.conn.Write([]byte{network.UserNodeJoin})
	if err != nil {
		return err
	}

	_, err = io.CopyN(user, sig.conn, 6)

	return err
}

func (lb *loadBalancer) nodeJoinHandler(node net.Conn) error {
	nodeJoinSignal := chanSignal{
		chanOpCode_nodeJoin,
		node,
		nil,
	}

	lb.connChan <- nodeJoinSignal
	if err := (<-lb.connChan).err; err != nil {
		return err
	}

	return nil
}
