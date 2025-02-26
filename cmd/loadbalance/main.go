package main

import (
	"errors"
	"flag"
	"io"
	"log/slog"
	"net"

	"github.com/hn275/distributed-storage/internal/loadbalance"
	"github.com/hn275/distributed-storage/internal/network"
)

var (
	serverAddr string
	logger     *slog.Logger = slog.Default()
)

func init() {
	flag.StringVar(&serverAddr, "addr", ":8000", "address to bind the load balancer")
	flag.Parse()
}

func main() {
	lbSrv, err := loadbalance.NewBalancer(
		network.ProtoTcp4,
		serverAddr,
		loadbalance.NewSimpleAlgo(),
	)
	if err != nil {
		panic(err)
	}

	defer lbSrv.Close()
	logger.Info(
		"node started, waiting for services.",
		"protocol", lbSrv.Addr().Network(),
		"address", lbSrv.Addr(),
	)

	var buf [256]byte
	for {
		conn, err := lbSrv.Accept()
		if err != nil {
			logger.Error("failed to accept new conn.",
				"peer", conn.RemoteAddr,
				"err", err,
			)
			continue
		}

		// handling the initial ping
		if _, err = conn.Read(buf[:]); err != nil {
			if errors.Is(err, io.EOF) {
				// peer disconnected, silent return
				return
			}

			logger.Error("failed to read from socket.",
				"peer", conn.RemoteAddr(),
				"err", err,
			)
			return
		}

		logger.Info("new connection.", "peer", conn.RemoteAddr())
		switch buf[0] {
		case network.DataNodeJoin:
			if err := lbSrv.NodeJoin(conn); err != nil {
				slog.Error("failed to join new node", "addr", conn.RemoteAddr())
				conn.Close()
				continue
			}

			go serveDataNode(conn)
			logger.Info("data node joined cluster.", "addr", conn.RemoteAddr())

		default:
			logger.Error("unsupported ping message type.", "msgtype", buf[0])
			closeConn(conn)
		}
	}
}

func serveDataNode(conn net.Conn) {
	defer func() {
		closeConn(conn)
		// TODO: need a way to remove this connection from the cluster map
		// when the node leaves the cluster/code exploded
	}()
}

func closeConn(conn net.Conn) {
	if err := conn.Close(); err != nil && !errors.Is(err, io.EOF) {
		logger.Error("failed to close connection",
			"peer", conn.RemoteAddr(),
			"err", err,
		)
	}

	logger.Info("connection closed", "peer", conn.RemoteAddr())
}
