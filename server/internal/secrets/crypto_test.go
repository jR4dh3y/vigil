package secrets

import "testing"

func TestEncryptDecryptPlain(t *testing.T) {
	enc, err := Encrypt("", "secret")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if enc != "plain:secret" {
		t.Fatalf("expected plain prefix, got %q", enc)
	}
	got, err := Decrypt("", enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != "secret" {
		t.Fatalf("got %q want secret", got)
	}
}

func TestEncryptDecryptAES(t *testing.T) {
	key := "test-secrets-key"
	enc, err := Encrypt(key, "drive-refresh-token")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if enc == "drive-refresh-token" || enc == "plain:drive-refresh-token" {
		t.Fatalf("expected ciphertext, got %q", enc)
	}
	got, err := Decrypt(key, enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != "drive-refresh-token" {
		t.Fatalf("got %q want drive-refresh-token", got)
	}
}

func TestEncryptEmpty(t *testing.T) {
	enc, err := Encrypt("key", "")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if enc != "" {
		t.Fatalf("expected empty, got %q", enc)
	}
}
