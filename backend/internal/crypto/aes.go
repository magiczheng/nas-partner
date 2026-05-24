package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"strings"
)

const encPrefix = "ENC:"

var derivedKey []byte

func SetKey(rawSecret string) {
	hash := sha256.Sum256([]byte(rawSecret))
	derivedKey = hash[:]
}

func EncryptSecret(plaintext string) (string, error) {
	if plaintext == "" || strings.HasPrefix(plaintext, encPrefix) {
		return plaintext, nil
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return encPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptSecret(encoded string) (string, error) {
	if encoded == "" || !strings.HasPrefix(encoded, encPrefix) {
		return encoded, nil
	}

	data, err := base64.StdEncoding.DecodeString(encoded[len(encPrefix):])
	if err != nil {
		return encoded, nil
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return encoded, nil
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return encoded, nil
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return encoded, nil
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return encoded, nil
	}

	return string(plaintext), nil
}
