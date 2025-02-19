package main

import (
	"errors"
	"flag"
	"io"
	"log/slog"
	"net"
	"os"

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
	soc, err := net.Listen(network.ProtoTcp4, serverAddr)
	if err != nil {
		logger.Error("failed to start node", "err", err)
		// TODO: define a set of error codes and semantics to share between binaries
		os.Exit(1)
		return
	}

	defer soc.Close()
	logger.Info(
		"node started, waiting for services.",
		"protocol", soc.Addr().Network(),
		"address", soc.Addr(),
	)

	// array buffer to hold all the data nodes in the cluster
	// TODO: Emily can define this data structure later
	dataNodes := make(map[net.Addr]net.Conn)
	go func() {
		for _, nodeConn := range dataNodes {
			closeConn(nodeConn)
		}
	}()

	var buf [256]byte
	for {
		conn, err := soc.Accept()
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
			dataNodes[conn.RemoteAddr()] = conn
			go serveDataNode(conn)
			logger.Info("data node joined cluster.", "addr", conn.RemoteAddr())

		default:
			logger.Error("unsupported ping message type.", "msgtype", buf[0])
			closeConn(conn)
		}
	}
}

func serveDataNode(conn net.Conn) {
	// TODO: need a way to remove this connection from the cluster map
	// when the node leaves the cluster/code exploded
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
