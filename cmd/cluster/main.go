package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/hn275/distributed-storage/internal"
	"github.com/hn275/distributed-storage/internal/config"
	"github.com/hn275/distributed-storage/internal/telemetry"
)

var wg *sync.WaitGroup

func main() {
	wg = new(sync.WaitGroup)

	// reading in config
	configPath := internal.EnvOrDefault("CONFIG_PATH", config.DefaultConfigPath)
	globConf, err := config.NewConfig(configPath)
	if err != nil {
		panic(err)
	}

	conf := &globConf.Cluster
	if conf.Capacity == 0 {
		panic("invalid capacity")
	}

	expName := globConf.Experiment.Name

	// parse load balancing address
	lbNodeAddr := fmt.Sprintf("127.0.0.1:%d", globConf.LoadBalancer.LocalPort)

	// telemetry
	filePath := "tmp/output/cluster/cluster-" + expName + ".csv"
	tel, err := telemetry.New(filePath, eventHeaders)
	if err != nil {
		panic(err)
	}

	defer tel.Done()

	// initialize cluster
	slog.Info(
		"initializing cluster.",
		"node-count", conf.Node,
		"load-balancer-addr", lbNodeAddr,
	)

	wg.Add(int(conf.Node))

	for nodeID := uint16(0); nodeID < conf.Node; nodeID++ {
		go func(nodeIndex uint16) {

			overHeadParam := time.Duration(0)
			if !globConf.Experiment.Homogeneous {
				rand.Seed(int64(nodeIndex))

				// For the sake of replicability and to ensure nodes are "heterogeneous" enough,
				// we keep the first 10 nodes hardcoded for the heterogeneous case.
				// If you want to change this behaviour, you need to modify these if statements
				if nodeIndex < 3 {
					overHeadParam = time.Millisecond * (10)
				} else if nodeIndex < 6 {
					overHeadParam = time.Millisecond * (500)
				} else if nodeIndex < 10 {
					overHeadParam = time.Millisecond * (900)
				} else {
					overHeadParam = time.Millisecond * time.Duration(rand.Int63n(globConf.Experiment.OverheadParam))
				}

				slog.Info("sleep timer", "v", overHeadParam)
			}

			node, err := nodeInitialize(lbNodeAddr, nodeID, tel, overHeadParam, conf.Capacity)
			if err != nil {
				slog.Error(
					"failed to initialize a data node.",
					"node-index", nodeID,
					"err", err,
				)
			}

			node.Listen()
		}(nodeID)
	}

	wg.Wait()
}
