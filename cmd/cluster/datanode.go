package main

import (
	"log/slog"
	"net"

	"github.com/hn275/distributed-storage/internal/network"
)

type DataNode struct {
	net.Conn
}

func nodeInitialize(lbAddr string) (*DataNode, error) {
	laddr, err := net.ResolveTCPAddr(network.ProtoTcp4, ":0") // randomize the port
	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveTCPAddr(network.ProtoTcp4, lbAddr)
	if err != nil {
		return nil, err
	}

	// dial and ping LB, notifying node type
	lbSoc, err := net.DialTCP(network.ProtoTcp4, laddr, raddr)
	if err != nil {
		return nil, err
	}

	if _, err := lbSoc.Write([]byte{network.DataNodeJoin}); err != nil {
		return nil, err
	}

	dataNode := &DataNode{lbSoc}

	return dataNode, nil
}

func (nodeConn *DataNode) Listen() {
	defer func() {
		// TODO: notify LB to join the node again
	}()

	var buf [6]byte
	// get a request from LB
	n, err := nodeConn.Read(buf[:])
	if err != nil {
		slog.Error("failed to read from LB.", "err", err)
		return
	}

	if n != 1 || buf[0] != network.UserNodeJoin {
		slog.Warn("invalid message from LB node.", "msg", string(buf[:]))
		return
	}

	// open a new port for user to dial
	soc, err := net.Listen(network.ProtoTcp4, ":0")
	if err != nil {
		slog.Error("failed to open new socket", "err", err)
		return
	}

	go serveUser(soc)

	// send to lb addr
	if _, err := nodeConn.Write([]byte(soc.Addr().String())); err != nil {
		panic(err)
	}
}

func serveUser(soc net.Listener) {
	defer soc.Close()

	slog.Info("waiting for user connection.", "addr", soc.Addr().String(), "protocol", soc.Addr().Network())

	user, err := soc.Accept()
	if err != nil {
		panic(err)
	}

	slog.Info("user connected.", "addr", user.RemoteAddr(), "protocol", user.RemoteAddr().Network())
}
