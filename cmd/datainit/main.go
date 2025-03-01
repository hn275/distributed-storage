package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/dustin/go-humanize"
	"lukechampine.com/blake3"
)

const (
	digestSize int = 32

	// i started using these
	bufferFileName string = "buf"
	fileCount      int    = 12
)

func makeFilePath(fileName string) string {
	return "tmp/data/" + fileName
}

type database struct {
	Xsmall  []string `json:"x-small"`
	Small   []string `json:"small"`
	Medium  []string `json:"medium"`
	Large   []string `json:"large"`
	Xlarge  []string `json:"x-large"`
	XXlarge []string `json:"xx-large"`
}

func main() {
	fileSizeOpts := [...]uint64{
		1 << 8,
		1 << 16,
		1 << 20,
		1 << 24,
		1 << 28,
		1 << 30,
	}

	fileSizeOptsLen := len(fileSizeOpts)

	const fileMode = os.O_RDWR | os.O_CREATE

	db := database{}

	for i := 0; i < fileCount; i++ {

		f, err := os.OpenFile(bufferFileName, fileMode, 0666)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		totalFileSize := fileSizeOpts[i%fileSizeOptsLen]

		if _, err := io.CopyN(f, rand.Reader, int64(totalFileSize)); err != nil {
			panic(err)
		}

		// resetting the cursor
		if _, err := f.Seek(0, 0); err != nil {
			panic(err)
		}

		h := blake3.New(digestSize, nil)
		if _, err := io.Copy(h, f); err != nil {
			panic(err)
		}

		digest := h.Sum(nil)
		fileName := makeFilePath(hex.EncodeToString(digest))

		if err := os.Rename(bufferFileName, fileName); err != nil {
			panic(err)
		}

		switch i % fileSizeOptsLen {
		case 0:
			db.Xsmall = append(db.Xsmall, fileName)
		case 1:
			db.Small = append(db.Small, fileName)
		case 2:
			db.Medium = append(db.Medium, fileName)
		case 3:
			db.Large = append(db.Large, fileName)
		case 4:
			db.Xlarge = append(db.Xlarge, fileName)
		case 5:
			db.XXlarge = append(db.XXlarge, fileName)
		default:
			panic("mod op out of index")
		}

		log.Println(fileName, "\t", humanize.Bytes(totalFileSize))
	}

	indexFileName := makeFilePath("file-index.json")

	f, err := os.OpenFile(indexFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}

	if err := json.NewEncoder(f).Encode(&db); err != nil {
		panic(err)
	}

	log.Println("created index json:", indexFileName)
}
