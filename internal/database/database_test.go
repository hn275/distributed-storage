package database

import (
	"encoding/hex"
	"testing"

	"github.com/hn275/distributed-storage/internal/crypto"
	"github.com/stretchr/testify/assert"
	"lukechampine.com/blake3"
)

func TestMakeFile(t *testing.T) {
	const fileSize = 0xff
	ciphertext, fileName, err := MakeFile(fileSize)
	assert.Nil(t, err)

	h := blake3.New(digestSize, nil)

	c := ciphertext[crypto.NonceSize : crypto.NonceSize+fileSize]
	_, err = h.Write(c)
	assert.Nil(t, err)

	assert.NotEqual(t, fileName, hex.EncodeToString(h.Sum(nil)))

	buf := make([]byte, 0xff)

	nonce := ciphertext[:crypto.NonceSize]

	err = crypto.Decrypt(
		buf[:0],
		crypto.DataNodeSecretKey[:],
		nonce,
		ciphertext[crypto.NonceSize:],
		crypto.UserPublicKey[:],
	)
	assert.Nil(t, err)

	h.Reset()
	_, err = h.Write(buf)

	assert.Nil(t, err)
	digest := h.Sum(nil)
	assert.Equal(t, fileName, hex.EncodeToString(digest))
}
