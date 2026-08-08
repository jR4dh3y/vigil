package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/nvr/nvr/server/internal/store"
)

var (
	// ErrSetupComplete reports that a user already exists and setup is finished.
	ErrSetupComplete = errors.New("setup already completed")
	// ErrInvalidAdminCredentials reports an unusable first-admin username/password pair.
	ErrInvalidAdminCredentials = errors.New("username required and password must be at least 8 characters")
)

// ValidateAdminCredentials validates a first-admin username/password pair. The
// username must be non-empty after trimming and the password at least 8 runes.
func ValidateAdminCredentials(username, password string) error {
	if strings.TrimSpace(username) == "" || utf8.RuneCountInString(password) < 8 {
		return ErrInvalidAdminCredentials
	}
	return nil
}

// CreateFirstAdmin creates the first admin user when the store has no users yet.
// It is safe for pre-listen bootstrap and repeated/concurrent invocation: when a
// user already exists it returns ErrSetupComplete without writing anything. The
// password is argon2id-hashed and never persisted or returned in plaintext.
func CreateFirstAdmin(ctx context.Context, q *store.Queries, username, password string) (*User, error) {
	username = strings.TrimSpace(username)
	if err := ValidateAdminCredentials(username, password); err != nil {
		return nil, err
	}

	count, err := q.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	if count > 0 {
		return nil, ErrSetupComplete
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := q.CreateUser(ctx, store.CreateUserParams{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: hash,
		Role:         RoleAdmin,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &User{ID: user.ID, Username: user.Username, Role: user.Role}, nil
}
