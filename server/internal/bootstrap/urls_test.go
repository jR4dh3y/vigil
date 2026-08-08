package bootstrap_test

import (
	"reflect"
	"testing"

	"github.com/nvr/nvr/server/internal/bootstrap"
)

func TestValidateURL(t *testing.T) {
	valid := []string{
		"https://dash.example.com",
		"http://lan:8080",
		"https://dash.example.com/path?x=1",
		" https://dash.example.com ",
	}
	for _, u := range valid {
		if err := bootstrap.ValidateURL(u); err != nil {
			t.Errorf("ValidateURL(%q) unexpected error: %v", u, err)
		}
	}

	invalid := []string{
		"",
		"   ",
		"dash.example.com",
		"://nohost",
		"ftp://example.com",
		"javascript:alert(1)",
	}
	for _, u := range invalid {
		if err := bootstrap.ValidateURL(u); err == nil {
			t.Errorf("ValidateURL(%q) expected error, got nil", u)
		}
	}
}

func TestOrigin(t *testing.T) {
	got, err := bootstrap.Origin("https://dash.example.com/some/path?q=1")
	if err != nil {
		t.Fatalf("Origin: %v", err)
	}
	if got != "https://dash.example.com" {
		t.Fatalf("Origin = %q, want https://dash.example.com", got)
	}
	if _, err := bootstrap.Origin("not a url"); err == nil {
		t.Fatal("Origin(non-url) expected error")
	}
}

func TestResolveURL(t *testing.T) {
	cases := []struct {
		env, db, want string
	}{
		{env: "https://env.example.com", db: "https://db.example.com", want: "https://env.example.com"},
		{env: "", db: "https://db.example.com", want: "https://db.example.com"},
		{env: "  ", db: "https://db.example.com", want: "https://db.example.com"},
		{env: "", db: "", want: ""},
	}
	for _, tc := range cases {
		if got := bootstrap.ResolveURL(tc.env, tc.db); got != tc.want {
			t.Errorf("ResolveURL(%q,%q) = %q, want %q", tc.env, tc.db, got, tc.want)
		}
	}
}

func TestCORSOrigins(t *testing.T) {
	// Configured HTTPS origin plus hosted dashboard origin are both allowed.
	got, err := bootstrap.CORSOrigins([]string{"https://client.example.com"}, "https://dash.example.com")
	if err != nil {
		t.Fatalf("CORSOrigins: %v", err)
	}
	want := []string{"https://client.example.com", "https://dash.example.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	// Non-local HTTP origin is rejected.
	if _, err := bootstrap.CORSOrigins([]string{"http://client.example.com"}, ""); err == nil {
		t.Fatal("expected error for non-local HTTP origin")
	}

	// Localhost HTTP origin is allowed (dev).
	got, err = bootstrap.CORSOrigins([]string{"http://localhost:5173"}, "")
	if err != nil {
		t.Fatalf("CORSOrigins localhost: %v", err)
	}
	if len(got) != 1 || got[0] != "http://localhost:5173" {
		t.Fatalf("localhost origin unexpected: %v", got)
	}

	// Duplicates are collapsed.
	got, err = bootstrap.CORSOrigins([]string{"https://a.example.com", "https://a.example.com"}, "https://a.example.com")
	if err != nil {
		t.Fatalf("CORSOrigins dedupe: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected deduped single origin, got %v", got)
	}
}
