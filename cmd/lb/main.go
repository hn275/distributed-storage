package main

import (
	"errors"
	"flag"
	"io"
	"log/slog"
	"net"
	"os"
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
	soc, err := net.Listen("tcp", serverAddr)
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

	for {
		conn, err := soc.Accept()
		if err != nil {
			logger.Error("failed to except new connection.",
				"peer", conn.RemoteAddr,
				"err", err,
			)
			continue
		}

		go serviceConn(conn)
	}
}

func serviceConn(conn net.Conn) {
	// disconnect
	defer func() {
		if err := conn.Close(); err != nil && !errors.Is(err, io.EOF) {
			logger.Error("failed to close connection",
				"peer", conn.RemoteAddr(),
				"err", err,
			)
			return
		}

		logger.Info("connection closed", "peer", conn.RemoteAddr())
	}()

	logger.Info("servicing connection.", "peer", conn.RemoteAddr())
	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err != nil {
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

		buf = buf[:n]
		logger.Info("message received.", "peer", conn.RemoteAddr(), "msg", string(buf))
	}
}
