package media

import (
	"testing"
	"time"
)

func TestMintAndValidateToken(t *testing.T) {
	store := NewTokenStore()
	token, expires, err := store.MintToken("cam-1", "cam_abc", DefaultTokenTTL)
	if err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if expires.Before(time.Now()) {
		t.Fatal("expires should be in the future")
	}

	if !store.ValidateAndConsume(token, "cam_abc") {
		t.Fatal("expected valid token for matching path")
	}
	// Reusable within TTL (HLS segments).
	if !store.ValidateAndConsume(token, "cam_abc") {
		t.Fatal("expected token reusable within TTL")
	}
	if store.ValidateAndConsume(token, "other_path") {
		t.Fatal("expected path mismatch to fail")
	}
	if store.ValidateAndConsume("nope", "cam_abc") {
		t.Fatal("expected unknown token to fail")
	}
}

func TestTokenExpiry(t *testing.T) {
	store := NewTokenStore()
	token, _, err := store.MintToken("cam-1", "cam_abc", 20*time.Millisecond)
	if err != nil {
		t.Fatalf("MintToken: %v", err)
	}
	time.Sleep(40 * time.Millisecond)
	if store.ValidateAndConsume(token, "cam_abc") {
		t.Fatal("expected expired token to fail")
	}
	if store.Len() != 0 {
		t.Fatalf("expected purged store, len=%d", store.Len())
	}
}

func TestPathName(t *testing.T) {
	got := PathName("550e8400-e29b-41d4-a716-446655440000")
	want := "cam_550e8400e29b41d4a716446655440000"
	if got != want {
		t.Fatalf("PathName: got %q want %q", got, want)
	}
}

func TestValidateAuthPasswordAndQuery(t *testing.T) {
	svc := NewService(Config{
		APIURL:    "http://127.0.0.1:9997",
		WebRTCURL: "http://127.0.0.1:8889",
		HLSURL:    "http://127.0.0.1:8888",
	}, nil)

	token, _, err := svc.tokens.MintToken("id", "cam_x", DefaultTokenTTL)
	if err != nil {
		t.Fatalf("mint: %v", err)
	}

	if !svc.ValidateAuth(AuthRequest{Action: "read", Path: "cam_x", Password: token}) {
		t.Fatal("password token should validate")
	}
	if !svc.ValidateAuth(AuthRequest{Action: "read", Path: "cam_x", Query: "token=" + token}) {
		t.Fatal("query token should validate")
	}
	if svc.ValidateAuth(AuthRequest{Action: "read", Path: "cam_x"}) {
		t.Fatal("empty creds should fail")
	}
}
