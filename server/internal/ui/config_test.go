package ui_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nvr/nvr/server/internal/ui"
)

func TestBuildDashboardURL(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "nvr.lan:8080"

	// Public URL is preferred for the server param.
	got := ui.BuildDashboardURL("https://dash.example.com/app", "https://nvr.public.com", r)
	want := "https://dash.example.com/app?server=https%3A%2F%2Fnvr.public.com"
	if got != want {
		t.Fatalf("with public URL = %q, want %q", got, want)
	}

	// Fallback to request origin when public URL is unset.
	got = ui.BuildDashboardURL("https://dash.example.com/app", "", r)
	want = "https://dash.example.com/app?server=http%3A%2F%2Fnvr.lan%3A8080"
	if got != want {
		t.Fatalf("request-origin fallback = %q, want %q", got, want)
	}

	// HTTPS request → https scheme in fallback.
	rTLS := httptest.NewRequest(http.MethodGet, "/", nil)
	rTLS.Host = "nvr.public.com"
	rTLS.TLS = &tls.ConnectionState{}
	got = ui.BuildDashboardURL("https://dash.example.com/app", "", rTLS)
	if got != "https://dash.example.com/app?server=https%3A%2F%2Fnvr.public.com" {
		t.Fatalf("TLS fallback = %q", got)
	}

	// Preserves existing query parameters (keys are sorted by Encode).
	got = ui.BuildDashboardURL("https://dash.example.com/app?x=1", "https://nvr.public.com", r)
	want = "https://dash.example.com/app?server=https%3A%2F%2Fnvr.public.com&x=1"
	if got != want {
		t.Fatalf("query preservation = %q, want %q", got, want)
	}

	// Empty hosted URL → empty result.
	if got := ui.BuildDashboardURL("", "https://nvr.public.com", r); got != "" {
		t.Fatalf("empty hosted URL should be empty, got %q", got)
	}
}
