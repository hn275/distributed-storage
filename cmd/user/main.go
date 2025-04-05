package main

import (
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hn275/distributed-storage/internal"
	"github.com/hn275/distributed-storage/internal/config"
	"github.com/hn275/distributed-storage/internal/crypto"
	"github.com/hn275/distributed-storage/internal/database"
	"github.com/hn275/distributed-storage/internal/network"
	"github.com/hn275/distributed-storage/internal/telemetry"
	"lukechampine.com/blake3"
)

const (
	simDeadLine     = 5 * time.Minute
	outputDir       = "tmp/output/user"
	timeStampFormat = "15:04:05.000"
)

var (
	lbNodeAddr     string
	shutdownSignal = [...]byte{network.ShutdownSig}
)

type ClientTimeData struct {
	duration  time.Duration
	size      int64
	timeStart time.Time
	timeEnd   time.Time
}

// Row implements telemetry.Record.
func (c *ClientTimeData) Row() []string {
	return []string{
		strconv.FormatInt(c.duration.Milliseconds(), 10), // duration
		strconv.FormatInt(c.size, 10),                    // size
		c.timeStart.Format(timeStampFormat),              // time-start
		strconv.FormatInt(c.timeStart.UnixNano(), 10),    //time-start(ns)
		c.timeEnd.Format(timeStampFormat),                // time-start
		strconv.FormatInt(c.timeEnd.UnixNano(), 10),      //time-end(ns)
	}
}

func main() {
	// make output dir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(err)
	}

	flag.StringVar(&lbNodeAddr, "lbaddr", "127.0.0.1:8000", "address of the loadbalancer")
	flag.Parse()

	slog.Info("Load balancing address.", "lbaddr", lbNodeAddr)

	// load config
	configPath := internal.EnvOrDefault("CONFIG_PATH", config.DefaultConfigPath)
	conf, err := config.NewConfig(configPath)
	if err != nil {
		panic(err)
	}

	// telemetry
	tel, err := telemetry.New(
		fmt.Sprintf("%s/client-%s.csv", outputDir, conf.Experiment.Name),
		[]string{
			"duration", "size", "time-start",
			"time-start(ns)", "time-end", "time-end(ns)"},
	)

	if err != nil {
		panic(err)
	}

	defer tel.Done()

	// get file addresses
	fileIndex, err := database.NewFileIndex()
	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)

	files := conf.User.GetFiles(fileIndex)

	clientIdx := 0
	// requesting files
	for fileName, freq := range files {
		slog.Info("requesting file.", "file-name", fileName, "freq", freq)
		wg.Add(freq)
		for range freq {
			// pass in global count variable
			go runSim(fileName, wg, clientIdx, conf.User.Interval, tel)
			// increment the global count variable
			clientIdx++
		}
	}

	wg.Wait()

	// send shutdown signal to load balancer
	// open socket to load balancer
	lbConn, err := net.Dial(network.ProtoTcp4, lbNodeAddr)
	if err != nil {
		panic(err)
	}

	defer lbConn.Close()

	if _, err := lbConn.Write(shutdownSignal[:]); err != nil {
		panic(err)
	}

	slog.Info("End of simulation")
}

func runSim(fileHash string, wg *sync.WaitGroup, clientIdx int, interval uint32, tel *telemetry.Telemetry) {
	defer wg.Done()

	// sleep for a random number of seconds in [0, interval]
	sleepDuration := time.Duration(rand.Float32()*float32(interval)) * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), simDeadLine)
	defer cancel()

	time.Sleep(sleepDuration)

	type Result struct {
		tele ClientTimeData
		err  error
	}

	doneChan := make(chan error, 1)

	go func() {
		timeStart := time.Now()

		fileSize, err := request(fileHash)
		if err != nil {
			doneChan <- err
			return
		}

		// Caputure request time for this client
		dur := time.Since(timeStart)
		timeEnd := timeStart.Add(dur)

		rec := &ClientTimeData{
			duration:  dur,
			size:      fileSize,
			timeStart: timeStart,
			timeEnd:   timeEnd,
		}

		tel.Collect(rec)

		slog.Info("file request.",
			"file-size", humanize.Bytes(uint64(rec.size)),
			"dur", rec.duration)

		doneChan <- nil
	}()

	select {
	case err := <-doneChan:
		if err != nil {
			slog.Error("request failed.",
				"file-hash", fileHash,
				"err", err)
		}
		return

	case <-ctx.Done():
		slog.Error("request timed out.", "client", clientIdx)
		return
	}
}

// request the file, returns (file size, error)
func request(fileHash string) (int64, error) {

	// open listener for data node
	// open a new port for user to dial
	soc, err := makeListener(network.ProtoTcp4, network.RandomLocalPort)
	if err != nil {
		return 0, err
	}

	defer soc.Close()

	// open socket to load balancer
	lbConn, err := net.Dial(network.ProtoTcp4, lbNodeAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to dial load balancer: %v", err)
	}

	defer lbConn.Close()

	ping := [16]byte{network.UserNodeJoin}
	if err := network.AddrToBytes(soc.Addr(), ping[1:]); err != nil {
		return 0, err
	}

	if _, err := lbConn.Write(ping[:]); err != nil {
		return 0, fmt.Errorf("failed ping load balancer: %v", err)
	}
	lbConn.Close()

	// datanode connects
	dataConn, err := soc.Accept()
	if err != nil {
		return 0, err
	}

	defer dataConn.Close()

	// sending file name + pub key
	// 64 bytes, 32 byte file hash, 32 byte pub key
	var buf [64]byte
	if _, err := hex.Decode(buf[:32], []byte(fileHash)); err != nil {
		return 0, fmt.Errorf("invalid file hash: %s", fileHash)
	}

	copy(buf[32:], crypto.UserPublicKey[:])

	if _, err := dataConn.Write(buf[:64]); err != nil {
		return 0, fmt.Errorf("failed to write to datanode; %v", err)
	}

	// slog.Info("file name + pub key sent", "addr", dataConn.RemoteAddr())

	// write responses to hasher
	h := blake3.New(crypto.DigestSize, crypto.UserPublicKey[:])

	byteCopied, err := io.Copy(h, dataConn)
	if err != nil {
		return 0, fmt.Errorf("failed to write to hasher; %v", err)
	}

	// hash the content then check the digest against the file name
	digest := h.Sum(nil)
	precomputedDigest := buf[:32]
	if !byteEqual(digest, precomputedDigest) {
		return 0, errors.New("file integrity violation")
	}

	return byteCopied, nil
}

func makeListener(protocol string, addr string) (net.Listener, error) {
	// reusing addr
	lc := net.ListenConfig{
		Control: func(_, _ string, c syscall.RawConn) error {
			var opErr error
			err := c.Control(func(fd uintptr) {
				opErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)

			})

			if err != nil {
				return err
			} else {
				return opErr
			}
		},
	}

	return lc.Listen(context.Background(), protocol, addr)
}

func byteEqual(buf1, buf2 []byte) bool {
	if len(buf1) != len(buf2) {
		return false
	}

	for i := range buf1 {
		if buf1[i] != buf2[i] {
			return false
		}
	}

	return true
}
