package camera

import "testing"

func TestEncryptDecryptPlain(t *testing.T) {
	enc, err := EncryptCredential("", "secret")
	if err != nil {
		t.Fatalf("EncryptCredential: %v", err)
	}
	if enc != "plain:secret" {
		t.Fatalf("expected plain prefix, got %q", enc)
	}
	got, err := DecryptCredential("", enc)
	if err != nil {
		t.Fatalf("DecryptCredential: %v", err)
	}
	if got != "secret" {
		t.Fatalf("got %q", got)
	}
}

func TestEncryptDecryptAES(t *testing.T) {
	key := "test-secrets-key"
	enc, err := EncryptCredential(key, "camera-pass")
	if err != nil {
		t.Fatalf("EncryptCredential: %v", err)
	}
	if enc == "" || enc == "camera-pass" || enc == "plain:camera-pass" {
		t.Fatalf("expected encrypted value, got %q", enc)
	}
	got, err := DecryptCredential(key, enc)
	if err != nil {
		t.Fatalf("DecryptCredential: %v", err)
	}
	if got != "camera-pass" {
		t.Fatalf("got %q", got)
	}
}

func TestEncryptEmpty(t *testing.T) {
	enc, err := EncryptCredential("key", "")
	if err != nil {
		t.Fatalf("EncryptCredential: %v", err)
	}
	if enc != "" {
		t.Fatalf("expected empty, got %q", enc)
	}
	got, err := DecryptCredential("key", "")
	if err != nil {
		t.Fatalf("DecryptCredential: %v", err)
	}
	if got != "" {
		t.Fatalf("got %q", got)
	}
}
