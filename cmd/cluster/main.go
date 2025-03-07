package main

import (
	"flag"
	"log/slog"
	"os"
	"sync"

	"github.com/hn275/distributed-storage/internal/config"
	"gopkg.in/yaml.v3"
)

var (
	lbNodeAddr string
)

func main() {
	flag.StringVar(&lbNodeAddr, "addr", "127.0.0.1:8000", "Load Balancing address")
	flag.Parse()

	config, err := config.NewClusterConfig("config.yml")
	if err != nil {
		panic(err)
	}

	configFd, err := os.OpenFile("config.yml", os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer configFd.Close()

	if err := yaml.NewDecoder(configFd).Decode(&config); err != nil {
		panic(err)
	}

	// initialize cluster
	slog.Info(
		"initializing cluster.",
		"node-count", config.Node,
		"load-balancer-addr", lbNodeAddr,
	)

	wg := new(sync.WaitGroup)
	wg.Add(int(config.Node))

	for nodeID := uint16(0); nodeID < config.Node; nodeID++ {
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
