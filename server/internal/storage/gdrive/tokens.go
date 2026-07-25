package gdrive

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nvr/nvr/server/internal/secrets"
	"github.com/nvr/nvr/server/internal/store"
	"golang.org/x/oauth2"
)

// Settings keys for Google Drive OAuth state and tokens.
const (
	KeyAccessToken       = "gdrive.access_token"
	KeyRefreshToken      = "gdrive.refresh_token"
	KeyTokenExpiry       = "gdrive.token_expiry"
	KeyTokenType         = "gdrive.token_type"
	KeyAccountEmail      = "gdrive.account_email"
	KeyConnectedAt       = "gdrive.connected_at"
	KeyFolderID          = "gdrive.folder_id"
	KeyOAuthState        = "gdrive.oauth_state"
	KeyOAuthStateExpires = "gdrive.oauth_state_expires"
)

var allTokenKeys = []string{
	KeyAccessToken,
	KeyRefreshToken,
	KeyTokenExpiry,
	KeyTokenType,
	KeyAccountEmail,
	KeyConnectedAt,
	KeyFolderID,
	KeyOAuthState,
	KeyOAuthStateExpires,
}

func (s *Service) getSetting(ctx context.Context, key string) (string, error) {
	if s == nil || s.q == nil {
		return "", fmt.Errorf("settings store is not configured")
	}
	row, err := s.q.GetSetting(ctx, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return row.Value, nil
}

func (s *Service) setSetting(ctx context.Context, key, value string) error {
	if s == nil || s.q == nil {
		return fmt.Errorf("settings store is not configured")
	}
	return s.q.UpsertSetting(ctx, store.UpsertSettingParams{Key: key, Value: value})
}

func (s *Service) deleteSetting(ctx context.Context, key string) error {
	if s == nil || s.q == nil {
		return fmt.Errorf("settings store is not configured")
	}
	return s.q.DeleteSetting(ctx, key)
}

func (s *Service) saveOAuthState(ctx context.Context, state string, expires time.Time) error {
	if err := s.setSetting(ctx, KeyOAuthState, state); err != nil {
		return err
	}
	return s.setSetting(ctx, KeyOAuthStateExpires, expires.UTC().Format(time.RFC3339))
}

func (s *Service) consumeOAuthState(ctx context.Context, state string) error {
	stored, err := s.getSetting(ctx, KeyOAuthState)
	if err != nil {
		return err
	}
	if stored == "" || subtle.ConstantTimeCompare([]byte(stored), []byte(state)) != 1 {
		return fmt.Errorf("invalid oauth state")
	}
	expStr, err := s.getSetting(ctx, KeyOAuthStateExpires)
	if err != nil {
		return err
	}
	// Single-use: clear before further checks so retries cannot reuse.
	if err := s.deleteSetting(ctx, KeyOAuthState); err != nil {
		return fmt.Errorf("consume oauth state: %w", err)
	}
	if err := s.deleteSetting(ctx, KeyOAuthStateExpires); err != nil {
		return fmt.Errorf("consume oauth state expiry: %w", err)
	}

	if expStr == "" {
		return fmt.Errorf("oauth state expired")
	}
	exp, err := time.Parse(time.RFC3339, expStr)
	if err != nil {
		return fmt.Errorf("oauth state expiry: %w", err)
	}
	if time.Now().UTC().After(exp) {
		return fmt.Errorf("oauth state expired")
	}
	return nil
}

func (s *Service) saveToken(ctx context.Context, tok *oauth2.Token, email string) error {
	if tok == nil {
		return fmt.Errorf("nil token")
	}
	if tok.RefreshToken == "" {
		return fmt.Errorf("refresh token required")
	}
	if strings.TrimSpace(s.secretsKey) == "" {
		return fmt.Errorf("NVR_SECRETS_KEY is required to store Google Drive tokens")
	}

	encAccess, err := secrets.Encrypt(s.secretsKey, tok.AccessToken)
	if err != nil {
		return fmt.Errorf("encrypt access token: %w", err)
	}
	encRefresh, err := secrets.Encrypt(s.secretsKey, tok.RefreshToken)
	if err != nil {
		return fmt.Errorf("encrypt refresh token: %w", err)
	}

	if err := s.setSetting(ctx, KeyAccessToken, encAccess); err != nil {
		return err
	}
	if err := s.setSetting(ctx, KeyRefreshToken, encRefresh); err != nil {
		return err
	}
	tokenType := tok.TokenType
	if tokenType == "" {
		tokenType = "Bearer"
	}
	if err := s.setSetting(ctx, KeyTokenType, tokenType); err != nil {
		return err
	}
	if !tok.Expiry.IsZero() {
		if err := s.setSetting(ctx, KeyTokenExpiry, tok.Expiry.UTC().Format(time.RFC3339)); err != nil {
			return err
		}
	}
	if email != "" {
		if err := s.setSetting(ctx, KeyAccountEmail, email); err != nil {
			return err
		}
	}
	// Preserve first-connect timestamp on token refresh (email empty).
	if email != "" {
		return s.setSetting(ctx, KeyConnectedAt, time.Now().UTC().Format(time.RFC3339))
	}
	existing, err := s.getSetting(ctx, KeyConnectedAt)
	if err != nil {
		return err
	}
	if existing != "" {
		return nil
	}
	return s.setSetting(ctx, KeyConnectedAt, time.Now().UTC().Format(time.RFC3339))
}

func (s *Service) loadToken(ctx context.Context) (*oauth2.Token, error) {
	encAccess, err := s.getSetting(ctx, KeyAccessToken)
	if err != nil {
		return nil, err
	}
	encRefresh, err := s.getSetting(ctx, KeyRefreshToken)
	if err != nil {
		return nil, err
	}
	if encRefresh == "" {
		return nil, nil
	}

	access, err := secrets.Decrypt(s.secretsKey, encAccess)
	if err != nil {
		return nil, fmt.Errorf("decrypt access token: %w", err)
	}
	refresh, err := secrets.Decrypt(s.secretsKey, encRefresh)
	if err != nil {
		return nil, fmt.Errorf("decrypt refresh token: %w", err)
	}

	tok := &oauth2.Token{
		AccessToken:  access,
		RefreshToken: refresh,
	}
	if typ, err := s.getSetting(ctx, KeyTokenType); err != nil {
		return nil, err
	} else if typ != "" {
		tok.TokenType = typ
	}
	if expStr, err := s.getSetting(ctx, KeyTokenExpiry); err != nil {
		return nil, err
	} else if expStr != "" {
		exp, err := time.Parse(time.RFC3339, expStr)
		if err != nil {
			return nil, fmt.Errorf("parse token expiry: %w", err)
		}
		tok.Expiry = exp
	}
	return tok, nil
}

func (s *Service) clearAllGDriveSettings(ctx context.Context) error {
	var first error
	for _, key := range allTokenKeys {
		if err := s.deleteSetting(ctx, key); err != nil && first == nil {
			first = err
		}
	}
	return first
}
