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

	"github.com/Jeffail/tunny"
	"github.com/hn275/distributed-storage/internal"
	"github.com/hn275/distributed-storage/internal/crypto"
	"github.com/hn275/distributed-storage/internal/database"
	"github.com/hn275/distributed-storage/internal/network"
	"github.com/hn275/distributed-storage/internal/telemetry"
)

const randomLocalPort = "127.0.0.1:0"

type dataNode struct {
	net.Conn

	id            uint16
	avgRT         float64       // average response time, in nanoseconds
	requestCtr    uint64        // num of requests served
	overHeadParam time.Duration // overhead in ms seconds, this is for sleep

	log  *slog.Logger
	tel  *telemetry.Telemetry
	mtx  *sync.Mutex
	pool *tunny.Pool
}

type request struct {
	msg       []byte
	timeStart time.Time
}

func (d *dataNode) Write(buf []byte) (int, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	n, err := d.Conn.Write(buf)
	return n, err
}

func nodeInitialize(
	lbAddr string, nodeID uint16, t *telemetry.Telemetry,
	overHeadParam time.Duration, capacity uint16,
) (*dataNode, error) {
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

	logger := slog.Default().With("node-id", nodeID)

	dataNode := &dataNode{
		Conn: lbSoc,

		id:            nodeID,
		avgRT:         0.0,
		requestCtr:    0,
		overHeadParam: overHeadParam,

		log:  logger,
		tel:  t,
		mtx:  new(sync.Mutex),
		pool: nil,
	}

	dataNode.pool = tunny.NewFunc(int(capacity), func(i any) any {
		return dataNode.handleUserJoin(i.(*request))
	})

	return dataNode, nil

}

func (d *dataNode) Listen() {
	defer wg.Done()
	defer d.Close()
	defer d.pool.Close()

	d.tel.Collect(&event{
		nodeID:       d.id,
		nodeOverhead: d.overHeadParam,
		eventType:    eventNodeOnline,
		peer:         "",
		timestamp:    time.Now(),
		duration:     0,
		size:         0,
		avgRT:        d.avgRT,
		activeReq:    d.requestCtr,
	})

	d.log.Info(
		"node online.",
		"addr", d.LocalAddr(),
	)

	for {
		buf := make([]byte, 16)
		// get a request from LB
		_, err := d.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				d.log.Info("load balancer disconnected.")
				break
			} else {
				d.log.Error("failed to read from LB.",
					"err", err)
			}
			return
		}

		switch buf[0] {
		case network.UserNodeJoin:
			wg.Add(1)
			go func() {
				defer wg.Done()

				req := &request{buf, time.Now()}
				if err := d.pool.Process(req); err != nil {
					d.log.Error("failed to service request.", "err", err.(error))
				}

				d.tel.Collect(&event{
					nodeID:       d.id,
					nodeOverhead: d.overHeadParam,
					eventType:    eventRequestRecv,
					peer:         "",
					timestamp:    req.timeStart,
					duration:     0,
					size:         0,
					avgRT:        d.avgRT,
					activeReq:    d.requestCtr,
				})
			}()

		default:
			d.log.Error("invalid requested service type.",
				"request", buf[0])

		}
	}

	d.tel.Collect(&event{
		nodeID:       d.id,
		nodeOverhead: d.overHeadParam,
		eventType:    eventNodeOffline,
		peer:         "",
		timestamp:    time.Now(),
		duration:     0,
		size:         0,
		avgRT:        d.avgRT,
		activeReq:    d.requestCtr,
	})
}

func (d *dataNode) handleUserJoin(req *request) error {
	if len(req.msg) != 16 {
		panic("handleUserJoin invalid buf size")
	}

	buf, ts := req.msg, &req.timeStart
	defer d.healthCheckReport(ts)

	time.Sleep(d.overHeadParam)

	// dial user's listener
	userAddr, err := network.BytesToAddr(buf[1:7])
	if err != nil {
		return err
	}

	user, err := net.DialTCP(network.ProtoTcp4, nil, userAddr.(*net.TCPAddr))
	if err != nil {
		return err
	}

	defer user.Close()

	// SERVING CLIENT
	d.log.Info("connected to user.", "addr", user.RemoteAddr())

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

	d.tel.Collect(&event{
		nodeID:       d.id,
		nodeOverhead: d.overHeadParam,
		eventType:    eventFileTransfer,
		peer:         peerLB,
		timestamp:    *ts,
		duration:     uint64(time.Since(*ts).Nanoseconds()),
		size:         uint64(n),
		avgRT:        d.avgRT,
		activeReq:    d.requestCtr,
	})

	return nil
}

func (d *dataNode) healthCheckReport(srvStartTime *time.Time) {
	ts := time.Now()

	// calculate the next average
	dur := float64(time.Since(*srvStartTime).Nanoseconds())
	d.avgRT = internal.CalcMovingAvg(d.requestCtr, d.avgRT, dur)
	d.requestCtr += 1

	// sends health check packet to LB
	buf := make([]byte, 16)
	buf[0] = network.HealthCheck

	bufWriter := bytes.NewBuffer(buf[1:1])
	if err := binary.Write(bufWriter, network.BinaryEndianess, d.avgRT); err != nil {
		panic(err) // TODO: log this
	}

	n, err := d.Write(buf)
	if err != nil {
		panic(err) // TODO: log this
	}

	d.tel.Collect(&event{
		nodeID:       d.id,
		nodeOverhead: d.overHeadParam,
		eventType:    eventHealthCheck,
		peer:         peerLB,
		timestamp:    ts,
		duration:     uint64(time.Since(ts).Nanoseconds()),
		size:         uint64(n),
		avgRT:        d.avgRT,
		activeReq:    d.requestCtr,
	})
}
