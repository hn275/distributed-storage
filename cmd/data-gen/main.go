package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/hn275/distributed-storage/internal/crypto"
	"github.com/hn275/distributed-storage/internal/database"
	"lukechampine.com/blake3"
)

var fileSizeOpts = [...]uint64{
	1 << 8,
	1 << 16,
	1 << 20,
	1 << 24,
	1 << 28,
	1 << 30,
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

	db := database.FileIndex{}

	for i, fileSize := range fileSizeOpts {
		fileBytes, fileName, err := makeFile(fileSize)
		if err != nil {
			panic(err)
		}

		// write to file
		filePath := database.AccessCluster.Append(fileName).String()
		fOut, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}

		if _, err := fOut.Write(fileBytes); err != nil {
			panic(err)
		}

		switch i {
		case 0:
			db.Xsmall = fileName
		case 1:
			db.Small = fileName
		case 2:
			db.Medium = fileName
		case 3:
			db.Large = fileName
		case 4:
			db.Xlarge = fileName
		case 5:
			db.XXlarge = fileName
		default:
			panic("index out of bound")
		}

		log.Printf("file created\t%s\t%s -> %s",
			fileName,
			humanize.Bytes(fileSize),
			humanize.Bytes(uint64(len(fileBytes))))
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

// return the (encrypted) content, and the file name, which is the hex encoded
// 32 byte hash of the non-encrypted content
func makeFile(fileSize uint64) ([]byte, string, error) {
	fileData := make([]byte, fileSize)
	if _, err := io.ReadFull(rand.Reader, fileData); err != nil {
		return nil, "", err
	}

	s, err := crypto.NewFileStream(
		crypto.DataNodeSecretKey[:], crypto.UserPublicKey[:],
	)
	if err != nil {
		return nil, "", err
	}

	h := blake3.New(crypto.DigestSize, crypto.UserPublicKey[:])

	buf := &bytes.Buffer{}
	if _, err := s.EncryptAndCopy(buf, bytes.NewReader(fileData), h); err != nil {
		return nil, "", err
	}

	fileName := hex.EncodeToString(h.Sum(nil))
	return buf.Bytes(), fileName, nil
}
