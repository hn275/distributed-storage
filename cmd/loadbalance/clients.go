package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"net"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hn275/distributed-storage/internal/algo"
	"github.com/hn275/distributed-storage/internal/network"
)

type clientMap struct{ sync.Map }

func (cx *clientMap) setClient(userConn net.Conn) {
	cx.Store(userConn.RemoteAddr().String(), userConn)
}

func (cx *clientMap) getClient(userAddr net.Addr) (net.Conn, bool) {
	v, ok := cx.LoadAndDelete(userAddr.String())
	if !ok {
		return nil, ok
	}
	return v.(net.Conn), ok
}

type dataNode struct {
	net.Conn
	wchan chan []byte
	id    uint16

	log      *slog.Logger
	avgRT    float64
	requests uint64
	index    int
}

// for debugging
func (d *dataNode) String() string {
	return fmt.Sprintf("[%d\t%f\t%d]", d.id, d.avgRT, d.requests)
}

// SetIndex implements algo.QueueNode.
func (d *dataNode) SetIndex(i int) {
	d.index = i
}

func (d *dataNode) Less(other algo.QueueNode) bool {
	o := other.(*dataNode)

	switch globConf.LoadBalancer.Algorithm {
	case algo.AlgoSimpleRoundRobin:
		return false // nop

	case algo.AlgoLeastResponseTime:
		if math.Abs(d.avgRT-o.avgRT) < math.SmallestNonzeroFloat64 {
			return d.requests < o.requests
		}
		return d.avgRT < o.avgRT

	case algo.AlgoLeastConnections:
		return d.requests < o.requests

	default:
		panic(fmt.Sprintf("not implemented [%s].", globConf.LoadBalancer.Algorithm))
	}
}

func makeDataNode(conn net.Conn, nodeID uint16) *dataNode {
	wchan := make(chan []byte, 100)

	logger := slog.Default().With("node-id", nodeID)

	dataNode := &dataNode{
		Conn:     conn,
		wchan:    wchan,
		id:       nodeID,
		log:      logger,
		avgRT:    0.0,
		requests: 0,
		index:    0,
	}
	go dataNode.listen()

	// write routine
	go func(wchan <-chan []byte) {
		for {
			buf := <-wchan
			if n, err := conn.Write(buf); err != nil {
				logger.Error("failed socket write.",
					"err", err,
					"len", humanize.Bytes(uint64(len(buf))))
			} else {
				logger.Info("packet sent.",
					"len", humanize.Bytes(uint64(n)))
			}
		}
	}(wchan)

	return dataNode
}

func (dn *dataNode) write(buf []byte) {
	dn.wchan <- buf
}

func (d *dataNode) listen() {
	for {
		buf := make([]byte, 16)
		_, err := d.Read(buf)
		if err != nil {
			d.log.Error("failed to read socket",
				"err", err)
			continue
		}

		switch buf[0] {

		case network.HealthCheck:
			go d.handleHealthCheck(buf)

		default:
			d.log.Error("unsupported message type", "type", buf[0])
		}
	}

}

func (d *dataNode) handleHealthCheck(buf []byte) {
	lbSrv.lock.Lock()

	ts := time.Now()
	defer lbSrv.lock.Unlock()

	// data node sends a health check message when it's done serving the client.
	// so the active requests is reduced by 1.
	d.requests -= 1

	bufReader := bytes.NewReader(buf[1:])
	err := binary.Read(bufReader, network.BinaryEndianess, &d.avgRT)
	if err != nil {
		d.log.Error("HealthCheck failed.", "err", err)
		return
	}

	if err := lbSrv.engine.Fix(d.index); err != nil {
		d.log.Error("failed priority queue fixes.", "err", err)
		return
	}

	lbSrv.tel.Collect(&event{
		eType:     eventHealthCheck,
		peer:      peerDataNode,
		peerID:    int32(d.id),
		timestamp: ts,
		duration:  time.Since(ts).Nanoseconds(),
		avgRT:     d.avgRT,
	})
}
