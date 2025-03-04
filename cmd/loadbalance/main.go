package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"

	"github.com/hn275/distributed-storage/internal/algo"
	"github.com/hn275/distributed-storage/internal/config"
	"github.com/hn275/distributed-storage/internal/network"
)

var (
	logger *slog.Logger = slog.Default()

	supportedAlgo = map[string]algo.LBAlgo{
		"simple-round-robin": &algo.RoundRobin{},
	}
)

func main() {
	conf, err := config.NewLB("config.yml")
	if err != nil {
		log.Fatalf("failed to read config. %v", err)
	}

	// initializing the lb
	var lbAlgo algo.LBAlgo
	lbAlgo, ok := supportedAlgo[conf.Algorithm]
	if !ok {
		log.Fatalf("unsupported algorithm: %s", conf.Algorithm)
	}

	lbAlgo.Initialize()
	log.Printf("load balancing algorithm: %s\n", conf.Algorithm)

	// open listening socket
	portStr := fmt.Sprintf(":%d", conf.LocalPort)
	soc, err := net.Listen(network.ProtoTcp4, portStr)

	if err != nil {
		log.Fatalf("failed to open listening socket: %W", err)
	}

	lbSrv := &loadBalancer{soc, lbAlgo, make(chan chanSignal, 128)}
	defer lbSrv.Close()
	go lbSrv.queryServer()

	logger.Info(
		"node started, waiting for services.",
		"protocol", lbSrv.Addr().Network(),
		"address", lbSrv.Addr(),
	)

	// serving
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

		default:
			logger.Error("unsupported ping message type.", "msgtype", buf[0])
			closeConn(conn)
		}
	}
}

type lbHandler func(net.Conn) error

// TODO: add telemetry
func handle(conn net.Conn, fn lbHandler) {
	if err := fn(conn); err != nil {
		logger.Error(
			"handler for peer returned an error.",
			"remote_addr", conn.RemoteAddr(),
			"err", err,
		)
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
