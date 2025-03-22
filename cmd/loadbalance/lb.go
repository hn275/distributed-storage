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

type lbHandler func(net.Conn, []byte) error

func handle(fn lbHandler, conn net.Conn, msg []byte) {
	if err := fn(conn, msg); err != nil {
		logger.Error(
			"handler for peer returned an error.",
			"remote_addr", conn.RemoteAddr(),
			"err", err,
		)
	}
}

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

		n, err := conn.Read(buf[:])
		if err != nil {
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
			go handle(lbSrv.nodeJoinHandler, conn, buf[:n])

		case network.UserNodeJoin:
			logger.Info("new user.", "remote_addr", conn.RemoteAddr())
			go handle(lbSrv.userJoinHandler, conn, nil)

		case network.ShutdownSig:
			return

		default:
			logger.Error("unsupported ping message type.", "msgtype", buf[0])
			closeConn(conn)
		}
	}
}

func (lb *loadBalancer) userJoinHandler(user net.Conn, _ []byte) error {
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

func (lb *loadBalancer) nodeJoinHandler(node net.Conn, msg []byte) error {
	if len(msg) != 3 {
		panic("protocol violation")
	}

	nodeId := network.BinaryEndianess.Uint16(msg[1:])

	dataNode := makeDataNode(node, nodeId)
	lb.lock.Lock()
	err := lb.engine.NodeJoin(dataNode)
	lb.lock.Unlock()

	dataNode.log.Info("new data node.", "remote_addr", node.RemoteAddr())
	return err
}
