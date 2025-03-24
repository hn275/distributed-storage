package database

import (
	"encoding/json"
	"os"

	"github.com/hn275/distributed-storage/internal/crypto"
)

const (
	Prefix        = "tmp/data/"
	AccessCluster = Path("cluster")
	AccessUser    = Path("user")
	fileOverhead  = crypto.TagSize + crypto.NonceSize
	digestSize    = 32
)

type Path string

func (p Path) String() string {
	return Prefix + string(p)
}

func (p Path) Append(path string) Path {
	if path[0] == '/' {
		return Path(string(p) + path)
	}

	return Path(string(p) + "/" + path)
}

// file addressing

type FileIndex struct {
	Xsmall  string `json:"x-small"`
	Small   string `json:"small"`
	Medium  string `json:"medium"`
	Large   string `json:"large"`
	Xlarge  string `json:"x-large"`
	XXlarge string `json:"xx-large"`
}

func NewFileIndex() (*FileIndex, error) {
	path := AccessUser.Append("file-index.json").String()

	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}

	fileIndex := new(FileIndex)
	err = json.NewDecoder(f).Decode(fileIndex)

	return fileIndex, err
}
