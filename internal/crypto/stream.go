package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	blockSize     = 0x200000 // 2mb block
	blockOverhead = NonceSize + TagSize
)

type FileStream struct {
	block  cipher.AEAD
	pubKey []byte
}

func NewFileStream(privateKey, pubKey []byte) (*FileStream, error) {
	block, err := chacha20poly1305.NewX(privateKey)
	if err != nil {
		return nil, err
	}

	f := &FileStream{block, pubKey}

	return f, nil
}

func (f *FileStream) DecryptAndCopy(dst io.Writer, src io.Reader) (int, error) {
	bytesDecrypted := 0

	bufSize := blockSize + blockOverhead
	fileBlock := make([]byte, bufSize)

	for {
		// read in a block
		n, err := src.Read(fileBlock)

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return bytesDecrypted, err
		}

		// decrypt block
		var (
			buf        = fileBlock[NonceSize:NonceSize]
			nonce      = fileBlock[:NonceSize]
			ciphertext = fileBlock[NonceSize:n]
		)

		buf, err = f.block.Open(buf, nonce, ciphertext, f.pubKey)
		if err != nil {
			return bytesDecrypted, err
		}

		// copy to dst
		n, err = dst.Write(buf)
		if err != nil {
			return bytesDecrypted, err
		}

		bytesDecrypted += n
	}

	return bytesDecrypted, nil
}

func (f FileStream) EncryptAndCopy(dst io.Writer, src io.Reader, hasher io.Writer) (int, error) {
	bytesEncrypted := 0

	// plaintext offset
	ptOffset := NonceSize + blockSize
	bufSize := ptOffset + TagSize
	fileBlock := make([]byte, bufSize)

	for {
		// read in a block
		plaintext := fileBlock[NonceSize:ptOffset]
		n, err := src.Read(plaintext)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return bytesEncrypted, err
		}

		// write plaintext to hasher
		plaintext = plaintext[:n]

		_, err = hasher.Write(plaintext)
		if err != nil {
			return bytesEncrypted, err
		}

		// make nonce
		_, err = io.ReadFull(rand.Reader, fileBlock[:NonceSize])
		if err != nil {
			return bytesEncrypted, err
		}

		// encrypt block
		nonce := fileBlock[:NonceSize]
		buf := f.block.Seal(plaintext[:0], nonce, plaintext, f.pubKey)

		// write to dst
		n, err = dst.Write(fileBlock[:len(buf)+NonceSize])
		if err != nil {
			return bytesEncrypted, err
		}

		bytesEncrypted += n
	}

	return bytesEncrypted, nil
}
