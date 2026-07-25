package camera

import "github.com/nvr/nvr/server/internal/secrets"

// EncryptCredential stores password material for cameras.
// When secretsKey is empty, the value is stored as plaintext with a "plain:" prefix (dev-friendly).
// When secretsKey is set, the value is AES-GCM encrypted with SHA-256(key) as the 32-byte key.
func EncryptCredential(secretsKey, plaintext string) (string, error) {
	return secrets.Encrypt(secretsKey, plaintext)
}

// DecryptCredential reverses EncryptCredential.
func DecryptCredential(secretsKey, stored string) (string, error) {
	return secrets.Decrypt(secretsKey, stored)
}
