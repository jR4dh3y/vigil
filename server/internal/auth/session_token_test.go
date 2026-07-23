package auth_test

import (
	"net/http"
	"testing"

	"github.com/nvr/nvr/server/internal/auth"
)

func TestSessionTokenFromRequest(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer abc123")
	tok, ok := auth.SessionTokenFromRequest(req)
	if !ok || tok != "abc123" {
		t.Fatalf("bearer: got %q %v", tok, ok)
	}

	req2, _ := http.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Session-Token", "xyz")
	tok, ok = auth.SessionTokenFromRequest(req2)
	if !ok || tok != "xyz" {
		t.Fatalf("header: got %q %v", tok, ok)
	}

	req3, _ := http.NewRequest(http.MethodGet, "/", nil)
	req3.AddCookie(&http.Cookie{Name: auth.CookieName, Value: "cookie-tok"})
	req3.Header.Set("Authorization", "Bearer bearer-tok")
	tok, ok = auth.SessionTokenFromRequest(req3)
	if !ok || tok != "cookie-tok" {
		t.Fatalf("cookie wins: got %q %v", tok, ok)
	}
}
