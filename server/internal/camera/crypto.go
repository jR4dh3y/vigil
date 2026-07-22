package camera

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const plainPrefix = "plain:"

// EncryptCredential stores password material for cameras.
// When secretsKey is empty, the value is stored as plaintext with a "plain:" prefix (dev-friendly).
// When secretsKey is set, the value is AES-GCM encrypted with SHA-256(key) as the 32-byte key.
func EncryptCredential(secretsKey, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if secretsKey == "" {
		return plainPrefix + plaintext, nil
	}

	key := sha256.Sum256([]byte(secretsKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptCredential reverses EncryptCredential.
func DecryptCredential(secretsKey, stored string) (string, error) {
	if stored == "" {
		return "", nil
	}
	if strings.HasPrefix(stored, plainPrefix) {
		return strings.TrimPrefix(stored, plainPrefix), nil
	}
	if secretsKey == "" {
		// Stored encrypted value but no key configured — cannot decrypt.
		return "", fmt.Errorf("secrets key required to decrypt credential")
	}

	raw, err := base64.StdEncoding.DecodeString(stored)
	if err != nil {
		return "", fmt.Errorf("decode credential: %w", err)
	}
	key := sha256.Sum256([]byte(secretsKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("credential ciphertext too short")
	}
	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt credential: %w", err)
	}
	return string(plain), nil
}
