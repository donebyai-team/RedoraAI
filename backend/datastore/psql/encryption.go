package psql

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

func (r *Database) encryptMessage(message string) (string, error) {
	byteMsg := []byte(message)
	block, err := aes.NewCipher(r.encryptionKey[:])
	if err != nil {
		return "", fmt.Errorf("could not create new cipher: %w", err)
	}

	cipherText := make([]byte, aes.BlockSize+len(byteMsg))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("read %d random bytes: %w", aes.BlockSize, err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], byteMsg)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func (r *Database) decryptMessage(message string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return "", fmt.Errorf("could not base64 decode: %w", err)
	}

	block, err := aes.NewCipher(r.encryptionKey[:])
	if err != nil {
		return "", fmt.Errorf("could not create new cipher: %v", err)
	}

	if len(cipherText) < aes.BlockSize {
		base := "the encrypted message is too short to be valid, "
		if len(message) == 0 {
			base += "the received message was empty and most probably the value was never set properly"
		} else {
			base += "have you properly encrypted it?"
		}

		return "", errors.New(base)
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}

func decodeEncryptionKey(key string) (out [aes.BlockSize]byte, err error) {
	decodedByteCount, err := hex.Decode(out[:], []byte(strings.TrimPrefix(key, "0x")))
	if err != nil {
		return out, fmt.Errorf("could not decode key: %w", err)
	}

	if decodedByteCount != aes.BlockSize {
		return out, fmt.Errorf("must be 16 bytes long (32 characters in hexadecimal form), got %d", decodedByteCount)
	}

	return out, nil
}
