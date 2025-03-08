package main

import (
	"flag"
	"log/slog"
	"sync"

	"github.com/hn275/distributed-storage/internal"
	"github.com/hn275/distributed-storage/internal/config"
)

var (
	lbNodeAddr string
)

func main() {
	flag.StringVar(&lbNodeAddr, "addr", "127.0.0.1:8000", "Load Balancing address")
	flag.Parse()

	conf, err := config.NewClusterConfig(internal.ConfigFilePath)
	if err != nil {
		panic(err)
	}

	// initialize cluster
	slog.Info(
		"initializing cluster.",
		"node-count", conf.Node,
		"load-balancer-addr", lbNodeAddr,
	)

	wg := new(sync.WaitGroup)
	wg.Add(int(conf.Node))

	for nodeID := uint16(0); nodeID < conf.Node; nodeID++ {
		go func(wg *sync.WaitGroup, nodeIndex uint16) {
			defer wg.Done()

			node, err := nodeInitialize(lbNodeAddr, nodeID)
			if err != nil {
				slog.Error(
					"failed to initialize a data node.",
					"node-index", nodeID,
					"err", err,
				)
			}

			slog.Info(
				"node online.",
				"node-index", nodeID,
				"addr", node.LocalAddr(),
			)
			node.Listen()
		}(wg, nodeID)
	}

	wg.Wait()
}
