package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/hn275/distributed-storage/internal/algo"
	"github.com/hn275/distributed-storage/internal/network"
	"github.com/hn275/distributed-storage/internal/telemetry"
)

const (
	eventUserJoin    = "user-joined"
	eventNodeJoin    = "node-joined"
	eventPortForward = "port-forward"
	eventHealthCheck = "health-check"

	peerUser     = "user"
	peerDataNode = "node"
)

var csvheaders = []string{
	"event-type", "peer", "node-id", "timestamp", "duration(ns)",
}

type event struct {
	eType     string
	peer      string
	peerID    int32
	timestamp time.Time
	duration  int64
}

func (e *event) Row() []string {
	return []string{
		e.eType,
		e.peer,
		fmt.Sprintf("%d", e.peerID),
		fmt.Sprintf("%d", e.timestamp.UnixNano()),
		fmt.Sprintf("%d", e.duration),
	}
}

type loadBalancer struct {
	net.Listener
	engine algo.LBAlgo
	lock   *sync.Mutex
	tel    *telemetry.Telemetry
}

func newLB(port int, algorithm algo.LBAlgo, tel *telemetry.Telemetry) (*loadBalancer, error) {
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
		tel:      tel,
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
	ts := time.Now()
	// port fowarding
	buf := [16]byte{network.UserNodeJoin}
	if err := network.AddrToBytes(user.RemoteAddr(), buf[1:7]); err != nil {
		panic(err)
	}

	// request for a data node
	lb.lock.Lock()
	defer lb.lock.Unlock()

	node, err := lb.engine.GetNode()
	if err != nil {
		return err
	}

	nodeQ := node.(*dataNode)
	nodeQ.requests += 1

	nodeQ.write(buf[:])
	cxMap.setClient(user)

	lb.engine.PutNode(nodeQ)

	lb.tel.Collect(&event{
		eType:     eventUserJoin,
		peer:      peerUser,
		peerID:    int32(nodeQ.id),
		timestamp: ts,
		duration:  time.Since(ts).Nanoseconds(),
	})
	return err
}

func (lb *loadBalancer) nodeJoinHandler(node net.Conn, msg []byte) error {
	ts := time.Now()
	if len(msg) != 3 {
		panic("protocol violation")
	}

	nodeId := network.BinaryEndianess.Uint16(msg[1:])

	dataNode := makeDataNode(node, nodeId)
	lb.lock.Lock()
	lb.engine.NodeJoin(dataNode)
	lb.lock.Unlock()

	dataNode.log.Info("new data node.", "remote_addr", node.RemoteAddr())
	lb.tel.Collect(&event{
		eType:     eventNodeJoin,
		peer:      peerDataNode,
		peerID:    int32(nodeId),
		timestamp: ts,
		duration:  time.Since(ts).Nanoseconds(),
	})
	return nil
}
