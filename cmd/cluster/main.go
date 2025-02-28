package main

import (
	"log/slog"
)

const (
	DataNodeCount int    = 5 // TODO: pull this in from a yaml file
	PortPrefix    int    = 9000
	LBNodeAddr    string = "127.0.0.1:8000"
)

func main() {
	nodes := make([]*dataNode, DataNodeCount)

	for i := 0; i < DataNodeCount; i++ {
		var err error
		nodes[i], err = nodeInitialize(LBNodeAddr)
		if err != nil {
			panic(err)
		}

		slog.Info("node joined cluster.", "addr", nodes[i].LocalAddr())
		go nodes[i].Listen()
	}

	select {}
}
