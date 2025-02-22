package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"lukechampine.com/blake3"
)

const (
	usage string = `usage: %s <file-size>
	file-size: in bytes, the value should be an unsigned 64 bit integer
`

	blockSize   uint64 = 1 << 12
	digestSize  int    = 32
	tmpFileName string = "tmp-data"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
		os.Exit(1)
	}

	totalFileSize, err := strconv.ParseUint(os.Args[1], 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	var hashKey []byte = nil // TODO: hard code this key
	h := blake3.New(digestSize, hashKey)

	f, err := os.OpenFile(tmpFileName, os.O_RDONLY|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	blocks := uint16(math.Ceil(float64(totalFileSize) / float64(blockSize)))
	var buf [blockSize]byte
	for i := uint16(0); i < blocks; i++ {
		var iterSize uint64
		if totalFileSize > blockSize {
			iterSize = blockSize
		} else {
			iterSize = totalFileSize
		}

		n, err := rand.Reader.Read(buf[:iterSize])
		if err != nil {
			panic(err)
		}

		// write to file
		if _, err := f.Write(buf[:n]); err != nil {
			panic(err)
		}

		// write to hasher
		if _, err := h.Write(buf[:n]); err != nil {
			panic(err)
		}
		totalFileSize -= uint64(n)
	}

	// TODO: rename the file
	fileName := hex.EncodeToString(h.Sum(nil))

	if err := os.Rename(tmpFileName, fileName); err != nil {
		panic(err)
	}

	fmt.Println("fileSize", totalFileSize, "fileName", fileName)
}
