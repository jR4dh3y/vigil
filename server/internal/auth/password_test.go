package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" || hash == "correct-horse-battery" {
		t.Fatalf("expected argon2id hash, got %q", hash)
	}

	ok, err := VerifyPassword("correct-horse-battery", hash)
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Fatal("expected password to match")
	}

	ok, err = VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword wrong: %v", err)
	}
	if ok {
		t.Fatal("expected password mismatch")
	}
}

func TestNewSessionToken(t *testing.T) {
	tok1, hash1, err := NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken: %v", err)
	}
	if tok1 == "" || hash1 == "" {
		t.Fatal("empty token or hash")
	}
	if hash1 != HashToken(tok1) {
		t.Fatal("hash mismatch")
	}

	tok2, hash2, err := NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken 2: %v", err)
	}
	if tok1 == tok2 || hash1 == hash2 {
		t.Fatal("tokens should be unique")
	}
}
