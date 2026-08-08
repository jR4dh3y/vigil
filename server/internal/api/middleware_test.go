package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newCORSTestHandler(origins ...string) http.Handler {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return CORSMiddleware(origins)(next)
}

func TestCORSConfiguredOriginNoCredentials(t *testing.T) {
	h := newCORSTestHandler("https://client.example.com")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://client.example.com")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://client.example.com" {
		t.Fatalf("allow-origin = %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, Authorization, X-Session-Token" {
		t.Fatalf("allow-headers = %q", got)
	}
	if got := rr.Header().Get("Access-Control-Expose-Headers"); got != "X-Session-Token" {
		t.Fatalf("expose-headers = %q", got)
	}
	if got := rr.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("vary = %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("configured origin must not get Allow-Credentials, got %q", got)
	}
}

func TestCORSLocalhostKeepsCredentials(t *testing.T) {
	h := newCORSTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("allow-origin = %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("localhost must keep Allow-Credentials, got %q", got)
	}
}

func TestCORSUnknownOriginGetsNothing(t *testing.T) {
	h := newCORSTestHandler("https://client.example.com")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unknown origin got allow-origin %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("unknown origin got credentials %q", got)
	}
}

func TestCORSPreflightConfigured(t *testing.T) {
	h := newCORSTestHandler("https://client.example.com")
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/x", nil)
	req.Header.Set("Origin", "https://client.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://client.example.com" {
		t.Fatalf("preflight allow-origin = %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("preflight must not allow credentials, got %q", got)
	}
}

func TestCORSPreflightUnknownOrigin(t *testing.T) {
	h := newCORSTestHandler("https://client.example.com")
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unknown preflight got allow-origin %q", got)
	}
}
