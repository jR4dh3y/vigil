package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/bootstrap"
	"github.com/nvr/nvr/server/internal/store"
)

func newSetupTestStore(t *testing.T) *store.Queries {
	t.Helper()
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return store.New(db)
}

func TestSetupPasswordSources(t *testing.T) {
	cases := []struct {
		name           string
		envPass        string
		passwordStdin  bool
		nonInteractive bool
		stdin          string
		prompt         string
		want           string
		wantErr        bool
	}{
		{name: "env wins", envPass: "envpass12", want: "envpass12"},
		{name: "stdin flag", passwordStdin: true, stdin: "stdinpass12\n", want: "stdinpass12"},
		{name: "prompt fallback", prompt: "promptpass12", want: "promptpass12"},
		{name: "noninteractive missing password", nonInteractive: true, wantErr: true},
		{name: "stdin overrides env (flag wins)", envPass: "envpass12", passwordStdin: true, stdin: "stdinpass12\n", want: "stdinpass12"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := setupOptions{
				EnvPassword:    tc.envPass,
				PasswordStdin:  tc.passwordStdin,
				NonInteractive: tc.nonInteractive,
				Stdin:          strings.NewReader(tc.stdin),
			}
			if tc.prompt != "" {
				o.ReadPassword = func() (string, error) { return tc.prompt, nil }
			}
			got, err := o.resolvePassword()
			if (err != nil) != tc.wantErr {
				t.Fatalf("resolvePassword err=%v wantErr=%v", err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Fatalf("password = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSetupFlagOverEnvPrecedence(t *testing.T) {
	o := setupOptions{
		Username:     "flaguser",
		EnvUsername:  "envuser",
		PublicURL:    "https://flag.example.com",
		EnvPublicURL: "https://env.example.com",
		HostedURL:    "",
		EnvHostedURL: "https://envdash.example.com",
	}
	if got := firstNonEmpty(o.Username, o.EnvUsername); got != "flaguser" {
		t.Fatalf("flag should override env username, got %q", got)
	}
	if got := firstNonEmpty(o.PublicURL, o.EnvPublicURL); got != "https://flag.example.com" {
		t.Fatalf("flag should override env public URL, got %q", got)
	}
	if got := firstNonEmpty(o.HostedURL, o.EnvHostedURL); got != "https://envdash.example.com" {
		t.Fatalf("env should fill absent hosted URL, got %q", got)
	}
}

func TestSetupPersistsAndCreatesAdmin(t *testing.T) {
	q := newSetupTestStore(t)
	ctx := context.Background()

	var out strings.Builder
	o := setupOptions{
		Username:    "admin",
		EnvPassword: "password123",
		PublicURL:   "https://nvr.public.com",
		HostedURL:   "https://dash.example.com",
		Stdin:       strings.NewReader(""),
		Stdout:      &out,
	}

	if err := runSetupWithDB(ctx, q, o); err != nil {
		t.Fatalf("runSetupWithDB: %v", err)
	}

	// Admin created.
	rows, err := q.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(rows) != 1 || rows[0].Username != "admin" {
		t.Fatalf("expected one admin, got %+v", rows)
	}

	// URLs persisted.
	if v, ok := settingValue(t, q, bootstrap.SettingPublicURL); !ok || v != "https://nvr.public.com" {
		t.Fatalf("publicUrl = %q ok=%v", v, ok)
	}
	if v, ok := settingValue(t, q, bootstrap.SettingHostedDashboardURL); !ok || v != "https://dash.example.com" {
		t.Fatalf("hostedDashboardUrl = %q ok=%v", v, ok)
	}

	// Output reports what was configured.
	if !strings.Contains(out.String(), "Admin username: admin") {
		t.Fatalf("output missing username: %q", out.String())
	}
	if !strings.Contains(out.String(), "https://dash.example.com") {
		t.Fatalf("output missing hosted URL: %q", out.String())
	}
}

func TestSetupRejectsAlreadyComplete(t *testing.T) {
	q := newSetupTestStore(t)
	ctx := context.Background()

	// First run completes setup.
	if _, err := auth.CreateFirstAdmin(ctx, q, "admin", "password123"); err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	o := setupOptions{
		Username:    "admin2",
		EnvPassword: "password123",
		Stdin:       strings.NewReader(""),
		Stdout:      &strings.Builder{},
	}
	if err := runSetupWithDB(ctx, q, o); !errors.Is(err, errSetupComplete) {
		t.Fatalf("expected errSetupComplete, got %v", err)
	}
}

func TestSetupUpdatesURLsWhenUsersExist(t *testing.T) {
	q := newSetupTestStore(t)
	ctx := context.Background()

	// Seed an existing admin.
	if _, err := auth.CreateFirstAdmin(ctx, q, "admin", "password123"); err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	// URL-only update: no username or password input supplied.
	var out strings.Builder
	o := setupOptions{
		HostedURL: "https://dash.example.com",
		PublicURL: "https://nvr.public.com",
		Stdin:     strings.NewReader(""),
		Stdout:    &out,
	}
	if err := runSetupWithDB(ctx, q, o); err != nil {
		t.Fatalf("URL-only update: %v", err)
	}

	// URLs updated without creating another user.
	rows, err := q.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("URL-only update must not create users, got %d", len(rows))
	}
	if v, ok := settingValue(t, q, bootstrap.SettingHostedDashboardURL); !ok || v != "https://dash.example.com" {
		t.Fatalf("hostedDashboardUrl = %q ok=%v", v, ok)
	}
	if v, ok := settingValue(t, q, bootstrap.SettingPublicURL); !ok || v != "https://nvr.public.com" {
		t.Fatalf("publicUrl = %q ok=%v", v, ok)
	}
	if !strings.Contains(out.String(), "Settings updated.") {
		t.Fatalf("output missing settings-updated marker: %q", out.String())
	}
}

func TestSetupRejectsInvalidURL(t *testing.T) {
	q := newSetupTestStore(t)
	o := setupOptions{
		Username:    "admin",
		EnvPassword: "password123",
		HostedURL:   "dash.example.com", // missing scheme
		Stdin:       strings.NewReader(""),
		Stdout:      &strings.Builder{},
	}
	if err := runSetupWithDB(context.Background(), q, o); err == nil {
		t.Fatal("expected invalid URL error")
	}
}

func settingValue(t *testing.T, q *store.Queries, key string) (string, bool) {
	t.Helper()
	row, err := q.GetSetting(context.Background(), key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false
		}
		t.Fatalf("GetSetting(%s): %v", key, err)
	}
	return row.Value, true
}
