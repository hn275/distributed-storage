package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sync"

	"github.com/hn275/distributed-storage/internal"
	"github.com/hn275/distributed-storage/internal/config"
	"github.com/hn275/distributed-storage/internal/telemetry"
)

func main() {

	// reading in config
	configPath := internal.EnvOrDefault("CONFIG_PATH", config.DefaultConfigPath)
	globConf, err := config.NewConfig(configPath)
	if err != nil {
		panic(err)
	}

	conf := &globConf.Cluster
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

	wg := new(sync.WaitGroup)
	wg.Add(int(conf.Node))

	const nsToMs = 1000000

	for nodeID := uint16(0); nodeID < conf.Node; nodeID++ {
		go func(wg *sync.WaitGroup, nodeIndex uint16) {

			overHeadParam := int64(0)
			if !globConf.Experiment.Homogeneous {
				overHeadParam = rand.Int63n(globConf.Experiment.OverheadParam) * nsToMs
			}

			node, err := nodeInitialize(lbNodeAddr, nodeID, tel, overHeadParam)
			if err != nil {
				slog.Error(
					"failed to initialize a data node.",
					"node-index", nodeID,
					"err", err,
				)
			}

			node.Listen(wg)
		}(wg, nodeID)
	}

	wg.Wait()
}
