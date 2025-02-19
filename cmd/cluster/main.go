package main

import (
	"fmt"
	"log/slog"
)

const (
	DataNodeCount int = 5 // TODO: pull this in from a yaml file
	PortPrefix    int = 9000
)

func main() {
	nodes := make([]*DataNode, DataNodeCount)

	for i := 0; i < DataNodeCount; i++ {
		var err error
		nodes[i], err = makeDataNode(fmt.Sprintf("0.0.0.0:%d", i+PortPrefix))
		if err != nil {
			panic(err)
		}

		if nodes[i] == nil {
			panic("alskdjflkdslkfj")
		}
		slog.Info("node joined cluster.", "addr", nodes[i].lbSoc.LocalAddr())

	}

	for _, node := range nodes {
		slog.Info("node left cluster.", "addr", node.lbSoc.LocalAddr())
		node.lbSoc.Close()
	}
}
