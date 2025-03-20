package main

import (
	"encoding/csv"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
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
// can put the struct here. 
type ClientTimeData struct {
    duration time.Duration
	size int64
}

var (
	lbNodeAddr string
	allClientTimeData []ClientTimeData

	shutdownSignal = [...]byte{network.ShutdownSig}
	userJoinSignal = [...]byte{network.UserNodeJoin}
)


func main() {

	flag.StringVar(&lbNodeAddr, "lbaddr", "127.0.0.1:8000", "address of the loadbalancer")
	flag.Parse()

	slog.Info("Load balancing address.", "lbaddr", lbNodeAddr)

	// load config
	userConf, err := config.NewUserConfig(internal.ConfigFilePath)
	if err != nil {
		panic(err)
	}

	// get file addresses
	fileIndex, err := database.NewFileIndex()
	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)
	
	files := userConf.GetFiles(fileIndex)
	numClients := 0
	
	for _, freq := range files {
		numClients = numClients + freq;
	}
	// dynamically allocate client time array
	allClientTimeData = make([]ClientTimeData, numClients)

	clientIdx := 0

	// requesting files
	for fileName, freq := range files {
		slog.Info("requesting file.", "file-name", fileName, "freq", freq)
		wg.Add(freq)
		for i := 0; i < freq; i++ {
			// pass in global count variable
			go runSim(fileName, wg, clientIdx)
			// increment the global count variable
			clientIdx++
		}
	}

	wg.Wait()

	// write client request time data to output file 
	writeResultsToFile("clientResults.csv")

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

func writeResultsToFile(filename string){
	dir := "client-telemetry-results"
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("error creating directory:", err)
	}

	filePath := filepath.Join(dir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		slog.Error("error creating file:", err)
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
		slog.Error("error writing csv:", err)
		return
	}
}

func runSim(fileHash string, wg *sync.WaitGroup, clientIdx int) {
	start := time.Now()

	fileSize, err := request(fileHash, wg)
	if err != nil {
		slog.Error(
			"failed to request file.",
			"file-name", fileHash,
			"err", err,
		)
		return
	}

	// Caputure request time for this client
	dur := time.Since(start)
	// Write time data to the global array for main thread to write to csv later
	allClientTimeData[clientIdx] = ClientTimeData{duration: dur, size: fileSize}

	slog.Info(
		"file request complete.",
		"file-name", fileHash,
		"file-size", humanize.Bytes(uint64(fileSize)),
		"duration", dur,
	)
}

// request the file, returns (file size, error)
func request(fileHash string, wg *sync.WaitGroup) (int64, error) {
	defer wg.Done()

	// open socket to load balancer
	lbConn, err := net.Dial(network.ProtoTcp4, lbNodeAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to dial load balancer: %v", err)
	}

	defer lbConn.Close()

	if _, err := lbConn.Write(userJoinSignal[:]); err != nil {
		return 0, fmt.Errorf("failed ping load balancer: %v", err)
	}

	// slog.Info("connected to LB.", "remote_addr", lbConn.RemoteAddr())

	// 64 bytes, 32 byte file hash, 32 byte pub key
	var buf [64]byte

	n, err := lbConn.Read(buf[:])
	if err != nil {
		return 0, fmt.Errorf("failed receive message from load balancer: %v", err)
	}

	if n != 6 {
		return 0, errors.New("protocol violation, expecting 6 bytes only.")
	}

	// dialing data node
	dataNodeAddr, err := network.BytesToAddr(buf[:n])
	if err != nil {
		return 0, fmt.Errorf("invalid network address: %v", buf[:n])
	}

	dataConn, err := net.DialTCP(network.ProtoTcp4, nil, dataNodeAddr.(*net.TCPAddr))
	if err != nil {
		return 0, fmt.Errorf("failed to dail data node: %v", err)
	}

	/*
		slog.Info(
			"data node connected.",
			"addr", dataConn.RemoteAddr(),
			"protocol", dataConn.RemoteAddr().Network(),
		)
	*/
	defer dataConn.Close()

	// sending file name + pub key
	if _, err := hex.Decode(buf[:32], []byte(fileHash)); err != nil {
		return 0, fmt.Errorf("invalid file hash: %s", fileHash)
	}

	copy(buf[32:], crypto.UserPublicKey[:])

	if _, err := dataConn.Write(buf[:64]); err != nil {
		return 0, fmt.Errorf("failed to write to datanode; %v", err)
	}

	// slog.Info("file name + pub key sent", "addr", dataConn.RemoteAddr())

	// write responses to hasher
	h := blake3.New(32, nil)

	byteCopied, err := io.Copy(h, dataConn)
	if err != nil {
		return 0, fmt.Errorf("failed to write to hasher; %v", err)
	}

	// hash the content then check the digest against the file name
	digest := h.Sum(nil)
	if !byteEqual(digest, buf[:32]) {
		return 0, errors.New("file integrity violation")
	}

	/*
		slog.Info(
			"file request completed.",
			"file-hash", fileHash,
			"file-size", humanize.Bytes(uint64(byteCopied)),
		)
	*/

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
