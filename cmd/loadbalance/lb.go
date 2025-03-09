package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/hn275/distributed-storage/internal/algo"
	"github.com/hn275/distributed-storage/internal/network"
)

type loadBalancer struct {
	net.Listener
	engine algo.LBAlgo
	lock   *sync.Mutex
}

func newLB(port int, algorithm algo.LBAlgo) (*loadBalancer, error) {
	// open listening socket
	portStr := fmt.Sprintf(":%d", port)
	soc, err := net.Listen(network.ProtoTcp4, portStr)

	if err != nil {
		return nil, err
	}

	lbSrv := &loadBalancer{soc, algorithm, new(sync.Mutex)}
	return lbSrv, nil
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
	lb.lock.Lock()
	node, err := lb.engine.GetNode()
	lb.lock.Unlock()

	if err != nil {
		return err
	}

	// port fowarding
	_, err = node.Write([]byte{network.UserNodeJoin})
	if err != nil {
		return err
	}

	_, err = io.CopyN(user, node, 6)

	return err
}

func (lb *loadBalancer) nodeJoinHandler(node net.Conn) error {
	lb.lock.Lock()
	err := lb.engine.NodeJoin(node)
	lb.lock.Unlock()
	return err
}
