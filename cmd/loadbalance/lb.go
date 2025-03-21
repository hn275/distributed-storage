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

	lbSrv := &loadBalancer{
		Listener: soc,
		engine:   algorithm,
		lock:     new(sync.Mutex),
	}
	return lbSrv, nil
}

// server handlers
// listener
func (lbSrv *loadBalancer) listen() {
	for {
		buf := [16]byte{}
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
	// request for a data node
	lb.lock.Lock()
	_node, err := lb.engine.GetNode()
	lb.lock.Unlock()

	if err != nil {
		return err
	}

	node := _node.(*dataNode)

	// port fowarding
	buf := [16]byte{network.UserNodeJoin}
	if err := network.AddrToBytes(user.RemoteAddr(), buf[1:7]); err != nil {
		panic(err)
	}

	node.write(buf[:])

	cxMap.setClient(user)

	return err
}

func (lb *loadBalancer) nodeJoinHandler(node net.Conn) error {
	dataNode := makeDataNode(node)
	lb.lock.Lock()
	err := lb.engine.NodeJoin(dataNode)
	lb.lock.Unlock()
	return err
}
