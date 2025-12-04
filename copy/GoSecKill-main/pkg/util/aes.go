package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"go.uber.org/zap"
)

var key []byte

func init() {
	key = make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		zap.L().Fatal("failed to generate key", zap.Error(err))
	}
}

func Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	cbc := cipher.NewCBCEncrypter(block, iv)
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	paddedText := append(plaintext, bytes.Repeat([]byte{byte(padding)}, padding)...)
	ciphertext := make([]byte, len(paddedText))
	cbc.CryptBlocks(ciphertext, paddedText)

	return append(iv, ciphertext...), nil
}

func Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	cbc := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	cbc.CryptBlocks(plaintext, ciphertext)

	padding := plaintext[len(plaintext)-1]
	return plaintext[:len(plaintext)-int(padding)], nil
}
