package crypto

import (
	"bytes"
	"crypto/rand"
	"io"
	"log"
	"testing"

	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/assert"
	"lukechampine.com/blake3"
)

func TestFileStream(t *testing.T) {
	fileSizes := []int{
		0x8000000,  // 128MB
		0x10000000, // 256MB
		0x20000000, // 512MB
		0x40000000, // 1GB

		// some odd bytes
		0x100b001,
		0x307da08,
		0x4000097,
	}

	for _, size := range fileSizes {
		log.Println("Testing file size", humanize.Bytes(uint64(size)))
		runTest(t, size)
	}
}

func runTest(t *testing.T, fileSize int) {
	// make random data
	randomData := make([]byte, fileSize)
	_, err := io.ReadFull(rand.Reader, randomData)
	assert.Nil(t, err)

	buf := bytes.NewReader(randomData)
	assert.Nil(t, err)

	// initialize h + streamer
	h := blake3.New(DigestSize, UserPublicKey[:])
	streamer, err := NewFileStream(DataNodeSecretKey[:], UserPublicKey[:])
	assert.Nil(t, err)

	_, err = h.Write(randomData)
	assert.Nil(t, err)
	expectedDigest := h.Sum(nil)

	// encrypt
	h.Reset()
	cBuf := &bytes.Buffer{}
	_, err = streamer.EncryptAndCopy(cBuf, buf, h)
	assert.Nil(t, err)

	digest1 := h.Sum(nil)

	// check for valid hash
	assert.True(t, bytesEqual(expectedDigest, digest1))

	// decrypt
	pBuf := &bytes.Buffer{}
	n, err := streamer.DecryptAndCopy(pBuf, cBuf)
	assert.Nil(t, err)
	assert.Equal(t, fileSize, n)
	assert.Equal(t, len(randomData), pBuf.Len())
	assert.True(t, bytesEqual(randomData, pBuf.Bytes()))

	// check for valid hash
	h.Reset()
	_, err = io.Copy(h, pBuf)
	assert.Nil(t, err)

	digest2 := h.Sum(nil)
	assert.True(t, bytesEqual(expectedDigest, digest2))
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
