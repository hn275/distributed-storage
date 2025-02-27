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
	algo := loadbalance.RoundRobin{}
	algo.Initialize()

	lbSrv, err := loadbalance.NewBalancer(
		network.ProtoTcp4,
		serverAddr,
		&algo,
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

		if _, err = conn.Read(buf[:]); err != nil {
			// peer disconnected, silent return
			if errors.Is(err, io.EOF) {
				continue
			}

			logger.Error("failed to read from socket.",
				"peer", conn.RemoteAddr(),
				"err", err,
			)
			return
		}

		switch buf[0] {
		case network.DataNodeJoin:
			logger.Info("new data node.", "remote_addr", conn.RemoteAddr())
			go handle(conn, lbSrv.NodeJoin)

		case network.UserNodeJoin:
			logger.Info("new user.", "remote_addr", conn.RemoteAddr())
			go handle(conn, lbSrv.UserHandler)

		default:
			logger.Error("unsupported ping message type.", "msgtype", buf[0])
			closeConn(conn)
		}
	}
}

type lbHandler func(net.Conn) error

func handle(conn net.Conn, fn lbHandler) {
	if err := fn(conn); err != nil {
		logger.Info("failed to serve user.", "addr", conn.RemoteAddr(), "err", err)
	}
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
