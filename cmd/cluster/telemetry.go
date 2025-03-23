package main

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hn275/distributed-storage/internal/telemetry"
)

type eventType string

type peerType string

const (
	peerLB   = peerType("load-balance")
	peerUser = peerType("user")

	eventPortForward  = eventType("port-forwarding")
	eventHealthCheck  = eventType("healthcheck")
	eventFileTransfer = eventType("file-transfer")
)

var eventHeaders = []string{
	"event-type", "peer", "timestamp", "duration(ns)", "bytes-transferred",
}

// telemetry
type event struct {
	nodeID    uint16
	eventType eventType
	peer      peerType
	timestamp time.Time
	duration  uint64 // in nanoseconds
	size      uint64 // in bytes
}

// Row implements telemetry.Record.
func (e *event) Row() []string {
	return []string{
		string(e.eventType),
		string(e.peer),
		e.timestamp.Format("15:04:05.000"),
		fmt.Sprintf("%d", e.duration),
		humanize.Bytes(e.size),
	}
}

func collectEvent(nodeID uint16, etype eventType, peer peerType, start time.Time, tel *telemetry.Telemetry, size uint64) {
	e := event{
		nodeID:    nodeID,
		peer:      peer,
		eventType: etype,
		timestamp: start,
		duration:  uint64(time.Since(start).Nanoseconds()),
		size:      size,
	}
	tel.Collect(&e)
}
