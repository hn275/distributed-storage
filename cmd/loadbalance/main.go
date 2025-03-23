package main

import (
	"errors"
	"io"
	"log"
	"log/slog"
	"net"

	"github.com/hn275/distributed-storage/internal"
	"github.com/hn275/distributed-storage/internal/algo"
	"github.com/hn275/distributed-storage/internal/config"
)

var (
	logger   *slog.Logger = slog.Default()
	globConf *config.Config

	supportedAlgo = map[string]algo.LBAlgo{
		algo.AlgoSimpleRoundRobin: &algo.RoundRobin{},
	}
)

func main() {
	var err error
	// reading in config file
	configPath := internal.EnvOrDefault("CONFIG_PATH", config.DefaultConfigPath)
	globConf, err = config.NewConfig(configPath)
	if err != nil {
		log.Fatalf("failed to read config. %v", err)
	}

	conf := &globConf.LoadBalancer

	// initializing the lb
	var lbAlgo algo.LBAlgo
	lbAlgo, ok := supportedAlgo[conf.Algorithm]
	if !ok {
		log.Fatalf("unsupported algorithm: [%s]", conf.Algorithm)
	}

	lbAlgo.Initialize()
	log.Printf("load balancing algorithm: %s\n", conf.Algorithm)

	lbSrv, err := newLB(int(conf.LocalPort), lbAlgo)
	if err != nil {
		log.Fatalf("failed to open listening socket: %W", err)
	}

	defer lbSrv.Close()

	logger.Info(
		"node started, waiting for services.",
		"protocol", lbSrv.Addr().Network(),
		"address", lbSrv.Addr(),
	)

	// serving
	lbSrv.listen()

	// TODO: write telemetry data out
	slog.Info("end of simulation")
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
