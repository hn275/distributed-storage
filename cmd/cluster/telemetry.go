package main

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
)

type eventType string

type peerType string

const (
	peerLB   = peerType("load-balance")
	peerUser = peerType("user")

	eventNodeOnline   = eventType("node-online")
	eventNodeOffline  = eventType("node-offline")
	eventPortForward  = eventType("port-forwarding")
	eventHealthCheck  = eventType("healthcheck")
	eventFileTransfer = eventType("file-transfer")
)

var eventHeaders = []string{
	"node-id",
	"performance-overhead(ns)",
	"event-type",
	"peer",
	"timestamp",
	"duration(ns)",
	"bytes-transferred",
}

// telemetry
type event struct {
	nodeID       uint16
	nodeOverhead int64
	eventType    eventType
	peer         peerType
	timestamp    time.Time
	duration     uint64 // in nanoseconds
	size         uint64 // in bytes
}

// Row implements telemetry.Record.
func (e *event) Row() []string {
	return []string{
		fmt.Sprintf("%d", e.nodeID),
		fmt.Sprintf("%d", e.nodeOverhead),
		string(e.eventType),
		string(e.peer),
		fmt.Sprintf("%d", e.timestamp.UnixNano()),
		fmt.Sprintf("%d", e.duration),
		humanize.Bytes(e.size),
	}
}
