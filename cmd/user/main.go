package main

import (
	"context"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hn275/distributed-storage/internal"
	"github.com/hn275/distributed-storage/internal/config"
	"github.com/hn275/distributed-storage/internal/crypto"
	"github.com/hn275/distributed-storage/internal/database"
	"github.com/hn275/distributed-storage/internal/network"
	"lukechampine.com/blake3"
)

const readDeadLine = 5 * time.Second

// can put the struct here.
type ClientTimeData struct {
	duration time.Duration
	size     int64
}

var (
	lbNodeAddr        string
	allClientTimeData []ClientTimeData

	shutdownSignal = [...]byte{network.ShutdownSig}
)

func main() {

	flag.StringVar(&lbNodeAddr, "lbaddr", "127.0.0.1:8000", "address of the loadbalancer")
	flag.Parse()

	slog.Info("Load balancing address.", "lbaddr", lbNodeAddr)

	// load config
	configPath := internal.EnvOrDefault("CONFIG_PATH", config.DefaultConfigPath)
	conf, err := config.NewConfig(configPath)
	if err != nil {
		panic(err)
	}

	// get file addresses
	fileIndex, err := database.NewFileIndex()
	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)

	files := conf.User.GetFiles(fileIndex)
	numClients := 0

	for _, freq := range files {
		numClients = numClients + freq
	}
	// dynamically allocate client time array
	allClientTimeData = make([]ClientTimeData, numClients)

	clientIdx := 0

	// requesting files
	for fileName, freq := range files {
		if freq == 0 {
			continue
		}

		slog.Info("requesting file.", "file-name", fileName, "freq", freq)

		wg.Add(freq)

		for range freq {
			// pass in global count variable
			go runSim(fileName, wg, clientIdx, conf.User.Interval)
			// increment the global count variable
			clientIdx++
		}

	}

	wg.Wait()

	// write client request time data to output file
	expName := conf.Experiment.Name
	writeResultsToFile("client-" + expName + ".csv")

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
}

func writeResultsToFile(filename string) {
	dir := "tmp/output/user"

	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("error creating directory",
			"err", err)
		return
	}

	filePath := filepath.Join(dir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		slog.Error("error creating file",
			"err", err)
		return
	}
	defer file.Close()

	records := [][]string{
		{"duration", "size"},
	}

	for _, data := range allClientTimeData {
		row := []string{
			strconv.FormatInt(data.duration.Milliseconds(), 10),
			strconv.FormatInt(data.size, 10),
		}
		records = append(records, row)
	}

	w := csv.NewWriter(file)
	w.WriteAll(records)

	if err := w.Error(); err != nil {
		slog.Error("error writing csv",
			"err", err)
		return
	}

	slog.Info("End of simulation")
}

type Result struct {
	tele ClientTimeData
	err  error
}

func runSim(fileHash string, wg *sync.WaitGroup, clientIdx int, interval uint32) {
	defer wg.Done()

	// sleep for a random number of seconds in [0, interval]
	sleepDuration := time.Duration(rand.Float32()*float32(interval)) * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	time.Sleep(sleepDuration)

	slog.Info("request sent.", "file-name", fileHash, "client-id", clientIdx)

	resultChan := make(chan Result, 10)

	go func() {
		start := time.Now()

		fileSize, err := request(fileHash)
		if err != nil {
			resultChan <- Result{
				err: err,
			}
			return
		}

		// Caputure request time for this client
		dur := time.Since(start)
		// Write time data to the global array for main thread to write to csv later
		resultChan <- Result{
			tele: ClientTimeData{duration: dur, size: fileSize},
			err:  nil,
		}
	}()

	select {
	case record := <-resultChan:
		if record.err != nil {
			slog.Error("request failed.",
				"file-hash", fileHash,
				"err", record.err)
		} else {
			allClientTimeData[clientIdx] = record.tele
			slog.Info("file request.",
				"file-size", humanize.Bytes(uint64(record.tele.size)),
				"dur", record.tele.duration)
		}

	case <-ctx.Done():
		allClientTimeData[clientIdx] = ClientTimeData{0, 0}
		slog.Error("request timed out.")
		return
	}

}

// request the file, returns (file size, error)
func request(fileHash string) (int64, error) {

	// open listener for data node
	// open a new port for user to dial
	soc, err := net.Listen(network.ProtoTcp4, network.RandomLocalPort)
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

	// write responses to hasher
	h := blake3.New(crypto.DigestSize, crypto.UserPublicKey[:])

	byteCopied, err := io.Copy(h, dataConn)
	if err != nil {
		return 0, fmt.Errorf("failed to write to hasher; %v", err)
	}

	// hash the content then check the digest against the file name
	digest := h.Sum(nil)
	if !byteEqual(digest, buf[:32]) {
		return 0, errors.New("file integrity violation")
	}

	return byteCopied, nil
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
