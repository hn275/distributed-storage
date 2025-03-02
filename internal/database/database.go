package database

import (
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/hn275/distributed-storage/internal/crypto"
	"lukechampine.com/blake3"
)

type Path string

const (
	Prefix        = "tmp/data/"
	AccessCluster = Path("cluster")
	AccessUser    = Path("user")
	fileOverhead  = crypto.OverHead + crypto.NonceSize
	digestSize    = 32
)

func (p Path) String() string {
	return Prefix + string(p)
}

func (p Path) Append(path string) Path {
	if path[0] == '/' {
		return Path(string(p) + path)
	}

	return Path(string(p) + "/" + path)
}

// return the (encrypted) content, and the file name, which is the hex encoded
// 32 byte hash of the non-encrypted content
func MakeFile(fileSize uint64) ([]byte, string, error) {
	buf := make([]byte, fileSize+fileOverhead)
	if _, err := io.ReadFull(rand.Reader, buf[:fileSize]); err != nil {
		return nil, "", err
	}

	var (
		key       = crypto.DataNodeSecretKey[:]
		aad       = crypto.UserPublicKey[:]
		nonce     = buf[:crypto.NonceSize]
		plaintext = buf[crypto.NonceSize : crypto.NonceSize+fileSize]
	)

	// hash the content for file name
	h := blake3.New(digestSize, nil)
	if _, err := h.Write(plaintext); err != nil {
		panic(err)
	}

	// encrypt the content
	dst := buf[crypto.NonceSize:crypto.NonceSize] // encrypt in place

	err := crypto.Encrypt(
		dst, key, nonce,
		plaintext, aad,
	)
	if err != nil {
		return nil, "", err
	}

	digest := h.Sum(nil)
	return buf, hex.EncodeToString(digest), err
}
