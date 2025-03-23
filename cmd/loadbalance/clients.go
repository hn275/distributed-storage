package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hn275/distributed-storage/internal/algo"
	"github.com/hn275/distributed-storage/internal/network"
)

var cxMap = new(clientMap)

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

// SetIndex implements algo.QueueNode.
func (d *dataNode) SetIndex(i int) {
	d.index = i
}

func (d *dataNode) Less(other algo.QueueNode) bool {
	switch globConf.LoadBalancer.Algorithm {
	case algo.AlgoSimpleRoundRobin:
		return false // nop

	case algo.AlgoLeastResponseTime:
		return d.avgRT < other.(*dataNode).avgRT

	case algo.AlgoLeastConnections:
		return d.requests < other.(*dataNode).requests

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
		buf := [16]byte{}
		_, err := d.Read(buf[:])
		if err != nil {
			d.log.Error("failed to read socket",
				"err", err)
			continue
		}

		switch buf[0] {
		case network.PortForwarding:
			go d.handlePortForward(buf[:])

		case network.HealthCheck:
			go d.handleHealthCheck(buf[:])

		default:
			d.log.Error("unsupported message type", "type", buf[0])
		}
	}

}

func (d *dataNode) handleHealthCheck(buf []byte) {
	lbSrv.lock.Lock()

	ts := time.Now()
	defer func() {
		lbSrv.lock.Unlock()
		lbSrv.tel.Collect(&event{
			eType:     eventHealthCheck,
			peer:      peerDataNode,
			peerID:    int32(d.id),
			timestamp: ts,
			duration:  time.Since(ts).Nanoseconds(),
		})
	}()

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
}

func (dn *dataNode) handlePortForward(buf []byte) {
	if len(buf) < 13 {
		panic("handlePortForward insufficient buf size")
	}

	clientAddr, err := network.BytesToAddr(buf[1:7])
	if err != nil {
		dn.log.Error("failed to marshal address bytes.")
		return
	}

	client, ok := cxMap.getClient(clientAddr)
	if !ok {
		dn.log.Error("client not found in map.", "client", clientAddr)
		return
	}

	defer client.Close()

	if _, err := client.Write(buf[7:13]); err != nil {
		dn.log.Error("failed to forward port to client.",
			"client", client.RemoteAddr())
	} else {
		dn.log.Info("address forwarded to client.",
			"client", client.RemoteAddr())
	}
}
