package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nvr/nvr/server/internal/store"
)

// Service implements password hashing and session lifecycle against the store.
type Service struct {
	Queries *store.Queries
}

// NewService constructs an auth Service.
func NewService(q *store.Queries) *Service {
	return &Service{Queries: q}
}

// CreateSession stores a new session for userID and returns the raw token and expiry.
func (s *Service) CreateSession(ctx context.Context, userID string) (token string, expires time.Time, err error) {
	token, hash, err := NewSessionToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expires = SessionExpiresAt()
	_, err = s.Queries.CreateSession(ctx, store.CreateSessionParams{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: FormatSQLiteTime(expires),
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("create session: %w", err)
	}
	return token, expires, nil
}

// UserFromToken looks up a session by raw token and returns the user if valid and unexpired.
func (s *Service) UserFromToken(ctx context.Context, token string) (*User, error) {
	row, err := s.Queries.GetSessionByTokenHash(ctx, HashToken(token))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get session: %w", err)
	}

	expires, err := ParseSQLiteTime(row.ExpiresAt)
	if err != nil {
		return nil, err
	}
	if time.Now().UTC().After(expires) {
		_ = s.Queries.DeleteSessionByTokenHash(ctx, row.TokenHash)
		return nil, nil
	}

	return &User{
		ID:       row.UserID,
		Username: row.Username,
		Role:     row.Role,
	}, nil
}

// DeleteSession removes the session matching the raw token (no-op if missing).
func (s *Service) DeleteSession(ctx context.Context, token string) error {
	if err := s.Queries.DeleteSessionByTokenHash(ctx, HashToken(token)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
