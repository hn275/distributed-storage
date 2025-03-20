package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"sync"

	"github.com/hn275/distributed-storage/internal/algo"
	"github.com/hn275/distributed-storage/internal/network"
)

type userMap struct{ sync.Map }

// stores the user connection in the map, writes the address to `keyBuf`
func (u *userMap) store(user net.Conn) {
	addr := user.RemoteAddr().String()
	u.Store(addr, user)
}

func (u *userMap) retrieve(userAddr net.Addr) (net.Conn, error) {
	key := userAddr.String()
	v, ok := u.LoadAndDelete(key)
	if !ok {
		return nil, fmt.Errorf("user address not found: [%s]", userAddr.String())
	}

	conn, ok := v.(net.Conn)
	if !ok {
		panic("invalid interface, expected `net.Conn`")
	}

	return conn, nil
}

type loadBalancer struct {
	net.Listener
	engine      algo.LBAlgo
	clusterLock *sync.Mutex
	users       *userMap
}

func newLB(port int, algorithm algo.LBAlgo) (*loadBalancer, error) {
	// open listening socket
	portStr := fmt.Sprintf(":%d", port)
	soc, err := net.Listen(network.ProtoTcp4, portStr)

	if err != nil {
		return nil, err
	}

	lbSrv := &loadBalancer{soc, algorithm, new(sync.Mutex), new(userMap)}
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

func (lb *loadBalancer) portForwardHandler(buf []byte) error {
	// get user connection
	userAddr, err := network.BytesToAddr(buf[1:7])
	if err != nil {
		return err
	}

	userConn, err := lb.users.retrieve(userAddr)
	if err != nil {
		return err
	}

	defer userConn.Close()

	// forward the address
	_, err = userConn.Write(buf[:13])

	return err
}

func (lb *loadBalancer) userJoinHandler(user net.Conn) error {
	// defer user.Close()
	buf := [1 + network.AddrBufSize]byte{}
	buf[0] = network.UserNodeJoin

	if err := network.AddrToBytes(user.RemoteAddr(), buf[1:]); err != nil {
		return err
	}

	lb.users.store(user)

	// request for a data node
	lb.clusterLock.Lock()
	node, err := lb.engine.GetNode()
	lb.clusterLock.Unlock()

	if err != nil {
		return err
	}

	// port fowarding
	_, err = node.Write(buf[:])
	return err
}

func (lb *loadBalancer) nodeJoinHandler(node net.Conn) error {
	lb.clusterLock.Lock()
	err := lb.engine.NodeJoin(node)
	lb.clusterLock.Unlock()
	go lb.nodeServe(node)
	return err
}

func (lb *loadBalancer) nodeServe(node net.Conn) {
	for {
		var buf [16]byte
		n, err := node.Read(buf[:])
		if err != nil {
			log.Println(err)
			return
		}

		msgType, msgBuf := buf[0], buf[1:n]
		switch msgType {

		case network.DataNodePort:
			if err := lb.portForwardHandler(buf[:n]); err != nil {
				log.Println(err)
			}

		case network.UserNodeJoin:
			if len(msgBuf) != 13 { // 6 bytes for each address
				log.Println("TODO: invalid msgBuf len")
				return
			}

			userAddr, err := network.BytesToAddr(buf[1:7])
			if err != nil {
				panic(err) // TODO: handle this error
			}

			user, err := lb.users.retrieve(userAddr)
			if err != nil {
				log.Printf("user not found: %v\n", buf[:6])
			}

			// changed message type to be DataNodeJoin, then forward to user
			buf[0] = network.DataNodeJoin
			n, err := user.Write(buf[:13])
			if err != nil {
				panic(err)
			}

			slog.Info(
				"port forwarded.",
				"node-id", node.RemoteAddr(),
				"user", user.RemoteAddr(),
				"bytes", n,
			)

		default:
			slog.Error(
				"unsupported message",
				"type", msgType,
				"message", buf[:],
			)
		}
	}
}
