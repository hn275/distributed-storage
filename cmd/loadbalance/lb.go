package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/hn275/distributed-storage/internal/algo"
	"github.com/hn275/distributed-storage/internal/network"
)

type userMap struct{ sync.Map }

// stores the user connection in the map, writes the address to `keyBuf`
func (u *userMap) store(keyBuf []byte, user net.Conn) error {
	addr := user.RemoteAddr()
	if err := network.AddrToBytes(user.RemoteAddr(), keyBuf); err != nil {
		return err
	}

	u.Store(addr, user)
	return nil
}

func (u *userMap) retrieve(userAddrBytes []byte) (net.Conn, bool) {
	userAddr, err := network.BytesToAddr(userAddrBytes)
	if err != nil {
		panic(err)
	}

	v, ok := u.LoadAndDelete(userAddr)
	if !ok {
		return nil, ok
	}

	conn, ok := v.(net.Conn)
	if !ok {
		panic("invalid interface, expected `net.Conn`")
	}

	return conn, ok
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

func (lb *loadBalancer) userJoinHandler(user net.Conn) error {
	// defer user.Close()
	buf := [1 + network.AddrBufSize]byte{}
	buf[0] = network.UserNodeJoin

	if err := lb.users.store(buf[1:], user); err != nil {
		return err
	}

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
	/*
		if err != nil {
			return err
		}

		// _, err = io.CopyN(user, node, 6)

		return err
	*/
}

func (lb *loadBalancer) nodeJoinHandler(node net.Conn) error {
	lb.clusterLock.Lock()
	err := lb.engine.NodeJoin(node)
	lb.clusterLock.Unlock()
	go lb.nodeServe(node)
	return err
}

func (lb *loadBalancer) nodeServe(node net.Conn) {
	var buf [16]byte
	for {
		n, err := node.Read(buf[:])
		if err != nil {
			log.Println(err)
			return
		}

		msgType, msgBuf := buf[0], buf[1:n]
		switch msgType {
		case network.UserNodeJoin:
			if len(msgBuf) != 12 { // 6 bytes for each address
				log.Println("TODO: invalid msgBuf len")
				return
			}

			user, ok := lb.users.retrieve(buf[:6])
			if !ok {
				log.Printf("user not found: %v\n", buf[:6])
			}

			// changed message type to be DataNodeJoin, then forward to user
			buf[0] = network.DataNodeJoin
			if _, err := user.Write(buf[:]); err != nil {
				panic(err)
			}

		default:
			log.Printf("unsupported msgType %v", msgType)
		}
	}
}
