package main

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/hn275/distributed-storage/internal/config"
	"gopkg.in/yaml.v3"
)

const (
	PortPrefix int    = 9000
	LBNodeAddr string = "127.0.0.1:8000" // TODO: pull this in from env
)

func main() {
	config, err := config.NewCluster("config.yml")
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

	wg := new(sync.WaitGroup)

	slog.Info(fmt.Sprintf("creating %d nodes.", config.Node))
	wg.Add(int(config.Node))

	for i := uint16(0); i < config.Node; i++ {
		go func(wg *sync.WaitGroup, nodeIndex uint16) {
			defer wg.Done()

			node, err := nodeInitialize(LBNodeAddr)
			if err != nil {
				slog.Error(
					"failed to initialize a data node.",
					"node-index", i,
					"err", err,
				)
			}

			slog.Info(
				"node online.",
				"node-index", i,
				"addr", node.LocalAddr(),
			)
			node.Listen()
		}(wg, i)
	}

	wg.Wait()
}
