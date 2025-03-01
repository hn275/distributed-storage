package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/hn275/distributed-storage/internal/database"
	"lukechampine.com/blake3"
)

const (
	digestSize     int    = 32
	bufferFileName string = "tmp/buf"
	fileCount      int    = 12
)

type dataStore struct {
	Xsmall  []string `json:"x-small"`
	Small   []string `json:"small"`
	Medium  []string `json:"medium"`
	Large   []string `json:"large"`
	Xlarge  []string `json:"x-large"`
	XXlarge []string `json:"xx-large"`
}

func main() {
	// created the needed directories
	dirsNeeded := [...]string{
		database.Prefix,
		database.AccessCluster.String(),
		database.AccessUser.String(),
	}

	for _, v := range dirsNeeded {
		if err := createDir(v); err != nil {
			log.Fatalf("failed to create %s:\n\t%W", v, err)
		}
	}

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

	db := dataStore{}

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
		fileName := database.
			AccessCluster.
			Append(hex.EncodeToString(digest)).
			String()

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

		log.Printf("file created\t%s\t%s", fileName, humanize.Bytes(totalFileSize))
	}

	indexFileName := database.AccessUser.Append("file-index.json").String()

	f, err := os.OpenFile(indexFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	if err := json.NewEncoder(f).Encode(&db); err != nil {
		panic(err)
	}

	log.Printf("file created\t%s\n", indexFileName)
}

func createDir(dir string) error {
	// check if directory exists
	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error checking if directory exists: %w", err)
		}
	} else {
		log.Printf("%s exists, deleting.\n", dir)
		err := os.RemoveAll(dir)
		if err != nil {
			return fmt.Errorf("failed to delete existing directory: %w", err)
		}
	}

	// Create new directory
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	log.Printf("dir created\t\t%s\n", dir)

	return nil
}
