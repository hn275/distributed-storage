package crypto

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	plaintext := "Hello, world!"

	buf := make([]byte, NonceSize+TagSize+len(plaintext))

	_, err := io.ReadFull(rand.Reader, buf[:NonceSize])
	assert.Nil(t, err)

	copy(buf[NonceSize:], []byte(plaintext))

	tmp := make([]byte, len(buf))
	copy(tmp, buf)

	err = Encrypt(
		buf[NonceSize:NonceSize],
		DataNodeSecretKey[:],
		buf[:NonceSize],
		buf[NonceSize:NonceSize+len(plaintext)],
		UserPublicKey[:],
	)
	assert.Nil(t, err)

	for i := 0; i < NonceSize; i++ {
		assert.Equal(t, tmp[i], buf[i])
	}

	for i := NonceSize; i < len(buf)-NonceSize; i++ {
		assert.NotEqual(t, tmp[i], buf[i])
	}
}

func TestDecrypt(t *testing.T) {
	// encrypt data
	plaintext := "Hello, world!"
	buf := make([]byte, NonceSize+TagSize+len(plaintext))
	_, _ = io.ReadFull(rand.Reader, buf[:NonceSize])
	copy(buf[NonceSize:], []byte(plaintext))

	_ = Encrypt(
		buf[NonceSize:NonceSize],
		DataNodeSecretKey[:],
		buf[:NonceSize],
		buf[NonceSize:NonceSize+len(plaintext)],
		UserPublicKey[:],
	)

	// decrypt data
	err := Decrypt(
		buf[:0],
		DataNodeSecretKey[:],
		buf[:NonceSize],
		buf[NonceSize:],
		UserPublicKey[:],
	)

	assert.Nil(t, err)
	assert.Equal(t, plaintext, string(buf[:len(plaintext)]))
}
