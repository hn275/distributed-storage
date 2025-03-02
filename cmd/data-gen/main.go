package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/hn275/distributed-storage/internal/database"
)

var fileSizeOpts = [...]uint64{
	1 << 8,
	1 << 16,
	1 << 20,
	1 << 24,
	1 << 28,
	1 << 30,
}

type dataStore struct {
	Xsmall  string `json:"x-small"`
	Small   string `json:"small"`
	Medium  string `json:"medium"`
	Large   string `json:"large"`
	Xlarge  string `json:"x-large"`
	XXlarge string `json:"xx-large"`
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

	db := dataStore{}

	for i, fileSize := range fileSizeOpts {
		fileBytes, fileName, err := database.MakeFile(fileSize)
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

		log.Printf("file created\t%s\t%s", fileName, humanize.Bytes(fileSize))
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
