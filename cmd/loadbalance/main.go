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

		// handling the initial ping
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

		logger.Info("new connection.", "peer", conn.RemoteAddr())

		switch buf[0] {
		case network.DataNodeJoin:
			if err := lbSrv.NodeJoin(conn); err != nil {
				slog.Error("failed to join new node", "addr", conn.RemoteAddr())
				continue
			}

			logger.Info("data node joined cluster.", "addr", conn.RemoteAddr())

		case network.UserNodeJoin:
			if err := handleUserConnection(conn); err != nil {
				logger.Info("failed to serve user.", "addr", conn.RemoteAddr())
			}

		default:
			logger.Error("unsupported ping message type.", "msgtype", buf[0])
			closeConn(conn)
		}
	}
}

func handleUserConnection(user net.Conn) error {
	logger.Info("user connected", "addr", user.RemoteAddr())
	return nil
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
