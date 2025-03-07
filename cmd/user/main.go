package main

import (
	"encoding/hex"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/hn275/distributed-storage/internal/config"
	"github.com/hn275/distributed-storage/internal/crypto"
	"github.com/hn275/distributed-storage/internal/database"
	"github.com/hn275/distributed-storage/internal/network"
	"lukechampine.com/blake3"
)

const LBNodeAddr string = "127.0.0.1:8000" // TODO: load this from env with default value

func main() {
	// load config
	_, err := config.NewUserConfig("config.yml")
	if err != nil {
		panic(err)
	}

	// get file addresses
	fileIndex, err := database.NewFileIndex()
	if err != nil {
		panic(err)
	}

	fileNames := [...]string{
		fileIndex.Xsmall,
		fileIndex.Small,
		fileIndex.Medium,
		fileIndex.Large,
		fileIndex.Xlarge,
		fileIndex.XXlarge,
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(fileNames))

	for _, fileName := range fileNames {
		go request(fileName, wg)
	}

	wg.Wait()
	slog.Info("done")
}

// TODO: general error handling, right now it panics
func request(fileHash string, wg *sync.WaitGroup) {
	defer wg.Done()

	slog.Info("requesting", "file", fileHash)

	// open socket to load balancer
	lbConn, err := net.Dial(network.ProtoTcp4, LBNodeAddr)
	if err != nil {
		panic(err)
	}

	defer lbConn.Close()

	if _, err := lbConn.Write([]byte{network.UserNodeJoin}); err != nil {
		panic(err)
	}

	slog.Info("connected to LB.", "remote_addr", lbConn.RemoteAddr())

	var buf [0xff]byte
	n, err := lbConn.Read(buf[:])
	if err != nil {
		panic(err)
	}

	if n != 6 {
		panic("protocol violation, expecting 6 bytes only.")
	}

	// dialing data node
	dataNodeAddr, err := network.BytesToAddr(buf[:n])
	if err != nil {
		panic(err)
	}

	dataConn, err := net.DialTCP(network.ProtoTcp4, nil, dataNodeAddr.(*net.TCPAddr))
	if err != nil {
		panic(err)
	}

	slog.Info(
		"data node connected.",
		"addr", dataConn.RemoteAddr(),
		"protocol", dataConn.RemoteAddr().Network(),
	)
	defer dataConn.Close()

	// sending file name + pub key
	if _, err := hex.Decode(buf[:32], []byte(fileHash)); err != nil {
		panic(err)
	}

	copy(buf[32:], crypto.UserPublicKey[:])

	if _, err := dataConn.Write(buf[:64]); err != nil {
		panic(err)
	}
	slog.Info("file name + pub key sent", "addr", dataConn.RemoteAddr())

	// write responses to file
	filePath := database.AccessUser.Append(fileHash).String()
	fileOut, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	defer fileOut.Close()

	byteRecv, err := io.Copy(fileOut, dataConn)
	if err != nil {
		panic(err)
	}

	// file validation, integrity check:
	// hash the content then check the digest against the file name
	if _, err := fileOut.Seek(0, 0); err != nil {
		panic(err)
	}

	h := blake3.New(32, nil)
	if _, err := io.Copy(h, fileOut); err != nil {
		panic(err)
	}

	digest := h.Sum(nil)

	if !byteEqual(digest, buf[:32]) {
		panic("file integrity violation")
	}

	slog.Info(
		"file request completed.",
		"file-hash", fileHash,
		"file-path", filePath,
		"file-size", humanize.Bytes(uint64(byteRecv)),
	)
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
