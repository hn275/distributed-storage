package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"github.com/hn275/distributed-storage/internal"
	"github.com/hn275/distributed-storage/internal/crypto"
	"github.com/hn275/distributed-storage/internal/database"
	"github.com/hn275/distributed-storage/internal/network"
	"github.com/hn275/distributed-storage/internal/telemetry"
)

const randomLocalPort = "127.0.0.1:0"

type dataNode struct {
	net.Conn
	id       uint16
	avgRT    float64 // average response time, in nanoseconds
	requests uint64  // num of requests served
	log      *slog.Logger
	tel      *telemetry.Telemetry
	mtx      *sync.Mutex
}

func makeDataNode(c net.Conn, id uint16, tel *telemetry.Telemetry) *dataNode {
	logger := slog.Default().With("node-id", id)
	dataNode := &dataNode{
		Conn:     c,
		id:       id,
		avgRT:    0.0,
		requests: 0,
		log:      logger,
		tel:      tel,
		mtx:      new(sync.Mutex),
	}

	return dataNode
}

func (d *dataNode) Write(buf []byte) (int, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	n, err := d.Conn.Write(buf)
	return n, err
}

func nodeInitialize(lbAddr string, nodeID uint16, t *telemetry.Telemetry) (*dataNode, error) {
	laddr, err := net.ResolveTCPAddr(network.ProtoTcp4, randomLocalPort)
	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveTCPAddr(network.ProtoTcp4, lbAddr)
	if err != nil {
		return nil, err
	}

	// dial and ping LB, notifying node type
	lbSoc, err := net.DialTCP(network.ProtoTcp4, laddr, raddr)
	if err != nil {
		return nil, err
	}

	ping := [3]byte{network.DataNodeJoin}
	network.BinaryEndianess.PutUint16(ping[1:], nodeID)

	if _, err := lbSoc.Write(ping[:]); err != nil {
		return nil, err
	}

	dataNode := makeDataNode(lbSoc, nodeID, t)

	return dataNode, nil
}

func (d *dataNode) Listen(t *telemetry.Telemetry) {
	defer d.Close()

	for {
		buf := make([]byte, 16)
		// get a request from LB
		_, err := d.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				d.log.Info("load balancer disconnected.")
			} else {
				d.log.Error("failed to read from LB.",
					"err", err)
			}
			return
		}

		switch buf[0] {
		case network.UserNodeJoin:
			go func() {
				if err := d.handleUserJoin(buf); err != nil {
					d.log.Error("failed to service UserNodeJoin",
						"err", err)
				}
			}()

		default:
			d.log.Error("invalid requested service type.",
				"request", buf[0])

		}
	}
}

func (d *dataNode) handleUserJoin(buf []byte) error {
	ts := time.Now()
	defer d.healthCheckReport(&ts)

	if len(buf) != 16 {
		panic("handleUserJoin invalid buf size")
	}

	userAddr, err := network.BytesToAddr(buf[1:7])
	if err != nil {
		return err
	}

	// open a new port for user to dial
	user, err := net.DialTCP(network.ProtoTcp4, nil, userAddr.(*net.TCPAddr))
	if err != nil {
		return err
	}

	defer user.Close()

	// SERVING CLIENT
	// collectEvent(d.id, eventPortForward, peerLB, ts, d.tel, uint64(n))

	d.log.Info("user connected.", "addr", user.RemoteAddr())

	// get file digest + pub key from user
	var fileBuf [64]byte
	if _, err := user.Read(fileBuf[:]); err != nil {
		return err
	}

	fileName := hex.EncodeToString(fileBuf[:32])
	filePath := database.AccessCluster.Append(fileName).String()

	// read + decrypt file
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}

	var (
		pubKey []byte = fileBuf[32:]
		secKey []byte = crypto.DataNodeSecretKey[:]
	)

	s, err := crypto.NewFileStream(secKey, pubKey)
	if err != nil {
		return err
	}

	n, err := s.DecryptAndCopy(user, file)
	if err != nil {
		return err
	}

	collectEvent(d.id, eventHealthCheck, peerLB, ts, d.tel, uint64(n))
	return nil
}

func (d *dataNode) healthCheckReport(srvStartTime *time.Time) {
	ts := time.Now()

	// calculate the next average
	dur := float64(time.Since(*srvStartTime).Nanoseconds())
	d.avgRT = internal.CalcMovingAvg(d.requests, d.avgRT, dur)
	d.requests += 1

	// sends health check packet to LB
	buf := make([]byte, 16)
	buf[0] = network.HealthCheck

	bufWriter := bytes.NewBuffer(buf[1:1])
	if err := binary.Write(bufWriter, network.BinaryEndianess, d.avgRT); err != nil {
		panic(err) // TODO: Pad + log this
	}

	n, err := d.Write(buf)
	if err != nil {
		panic(err) // TODO: log this
	}

	collectEvent(d.id, eventHealthCheck, peerLB, ts, d.tel, uint64(n))
}
