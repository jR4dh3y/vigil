package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/store"
)

func newTestStore(t *testing.T) *store.Queries {
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

func TestValidateAdminCredentials(t *testing.T) {
	cases := []struct {
		name    string
		user    string
		pass    string
		wantErr bool
	}{
		{name: "valid", user: "admin", pass: "password123", wantErr: false},
		{name: "whitespace username", user: "  admin  ", pass: "password123", wantErr: false},
		{name: "empty username", user: "", pass: "password123", wantErr: true},
		{name: "blank username", user: "   ", pass: "password123", wantErr: true},
		{name: "short password", user: "admin", pass: "1234567", wantErr: true},
		{name: "empty password", user: "admin", pass: "", wantErr: true},
		{name: "multibyte password counts runes", user: "admin", pass: "€€€€€€€€", wantErr: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := auth.ValidateAdminCredentials(tc.user, tc.pass)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateAdminCredentials(%q,%q) err=%v wantErr=%v", tc.user, tc.pass, err, tc.wantErr)
			}
		})
	}
}

func TestCreateFirstAdmin(t *testing.T) {
	q := newTestStore(t)
	ctx := context.Background()

	user, err := auth.CreateFirstAdmin(ctx, q, "admin", "password123")
	if err != nil {
		t.Fatalf("CreateFirstAdmin: %v", err)
	}
	if user.Username != "admin" || user.Role != "admin" {
		t.Fatalf("unexpected user: %+v", user)
	}

	// No plaintext password anywhere in the stored row.
	row, err := q.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if strings.Contains(row.PasswordHash, "password123") {
		t.Fatal("password persisted in plaintext")
	}

	// Second call is inert: setup is complete.
	if _, err := auth.CreateFirstAdmin(ctx, q, "admin2", "password123"); !errors.Is(err, auth.ErrSetupComplete) {
		t.Fatalf("expected ErrSetupComplete, got %v", err)
	}

	// Weak credentials are rejected before any write.
	if _, err := auth.CreateFirstAdmin(ctx, q, "x", "short"); !errors.Is(err, auth.ErrInvalidAdminCredentials) {
		t.Fatalf("expected ErrInvalidAdminCredentials, got %v", err)
	}
}

func TestCreateFirstAdminTrimsUsername(t *testing.T) {
	q := newTestStore(t)
	user, err := auth.CreateFirstAdmin(context.Background(), q, "  admin  ", "password123")
	if err != nil {
		t.Fatalf("CreateFirstAdmin: %v", err)
	}
	if user.Username != "admin" {
		t.Fatalf("username not trimmed: %q", user.Username)
	}
}
