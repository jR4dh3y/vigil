package gdrive

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
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
	KeyAccessToken   = "gdrive.access_token"
	KeyRefreshToken  = "gdrive.refresh_token"
	KeyTokenExpiry   = "gdrive.token_expiry"
	KeyTokenType     = "gdrive.token_type"
	KeyAccountEmail  = "gdrive.account_email"
	KeyConnectedAt   = "gdrive.connected_at"
	KeyFolderID      = "gdrive.folder_id"
	KeyOAuthStates   = "gdrive.oauth_states"
	KeyOAuthClientID = "gdrive.oauth_client_id"
	KeyOAuthSecret   = "gdrive.oauth_client_secret"
	KeyOAuthRedirect = "gdrive.oauth_redirect_url"
)

var allTokenKeys = []string{
	KeyAccessToken,
	KeyRefreshToken,
	KeyTokenExpiry,
	KeyTokenType,
	KeyAccountEmail,
	KeyConnectedAt,
	KeyFolderID,
	KeyOAuthStates,
}

const maxPendingOAuthStates = 16

type oauthStateEntry struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type connectionMetadata struct {
	AccountEmail string
	ConnectedAt  time.Time
}

func (s *Service) getSetting(ctx context.Context, key string) (string, error) {
	if s == nil || s.q == nil {
		return "", fmt.Errorf("settings store is not configured")
	}
	return getSetting(ctx, s.q, key)
}

func getSetting(ctx context.Context, q *store.Queries, key string) (string, error) {
	row, err := q.GetSetting(ctx, key)
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

func (s *Service) saveOAuthState(ctx context.Context, state string, expires time.Time) error {
	if s == nil || s.q == nil {
		return fmt.Errorf("settings store is not configured")
	}
	return store.InTransaction(ctx, s.q, func(q *store.Queries) error {
		states, err := loadOAuthStates(ctx, q)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		active := make([]oauthStateEntry, 0, len(states)+1)
		for _, pending := range states {
			if pending.ExpiresAt.After(now) {
				active = append(active, pending)
			}
		}
		active = append(active, oauthStateEntry{
			Value:     state,
			ExpiresAt: expires.UTC(),
		})
		if len(active) > maxPendingOAuthStates {
			active = active[len(active)-maxPendingOAuthStates:]
		}
		return saveOAuthStates(ctx, q, active)
	})
}

func (s *Service) consumeOAuthState(ctx context.Context, state string) error {
	if s == nil || s.q == nil {
		return fmt.Errorf("settings store is not configured")
	}

	var stateErr error
	err := store.InTransaction(ctx, s.q, func(q *store.Queries) error {
		states, err := loadOAuthStates(ctx, q)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		remaining := make([]oauthStateEntry, 0, len(states))
		matched := false
		for _, pending := range states {
			isMatch := subtle.ConstantTimeCompare([]byte(pending.Value), []byte(state)) == 1
			if isMatch {
				matched = true
				if !pending.ExpiresAt.After(now) {
					stateErr = fmt.Errorf("oauth state expired")
				}
				continue
			}
			if pending.ExpiresAt.After(now) {
				remaining = append(remaining, pending)
			}
		}
		if !matched {
			stateErr = fmt.Errorf("invalid oauth state")
		}
		return saveOAuthStates(ctx, q, remaining)
	})
	if err != nil {
		return err
	}
	return stateErr
}

func loadOAuthStates(ctx context.Context, q *store.Queries) ([]oauthStateEntry, error) {
	raw, err := getSetting(ctx, q, KeyOAuthStates)
	if err != nil || raw == "" {
		return nil, err
	}
	var states []oauthStateEntry
	if err := json.Unmarshal([]byte(raw), &states); err != nil {
		return nil, fmt.Errorf("decode oauth states: %w", err)
	}
	return states, nil
}

func saveOAuthStates(ctx context.Context, q *store.Queries, states []oauthStateEntry) error {
	if len(states) == 0 {
		return q.DeleteSetting(ctx, KeyOAuthStates)
	}
	raw, err := json.Marshal(states)
	if err != nil {
		return fmt.Errorf("encode oauth states: %w", err)
	}
	return q.UpsertSetting(ctx, store.UpsertSettingParams{
		Key:   KeyOAuthStates,
		Value: string(raw),
	})
}

func (s *Service) saveToken(
	ctx context.Context,
	tok *oauth2.Token,
	metadata *connectionMetadata,
) error {
	if s == nil || s.q == nil {
		return fmt.Errorf("settings store is not configured")
	}
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

	tokenType := tok.TokenType
	if tokenType == "" {
		tokenType = "Bearer"
	}

	return store.InTransaction(ctx, s.q, func(q *store.Queries) error {
		settings := []store.UpsertSettingParams{
			{Key: KeyAccessToken, Value: encAccess},
			{Key: KeyRefreshToken, Value: encRefresh},
			{Key: KeyTokenType, Value: tokenType},
		}
		if !tok.Expiry.IsZero() {
			settings = append(settings, store.UpsertSettingParams{
				Key:   KeyTokenExpiry,
				Value: tok.Expiry.UTC().Format(time.RFC3339),
			})
		}
		if metadata != nil {
			connectedAt := metadata.ConnectedAt
			if connectedAt.IsZero() {
				connectedAt = time.Now().UTC()
			}
			settings = append(settings,
				store.UpsertSettingParams{
					Key:   KeyAccountEmail,
					Value: strings.TrimSpace(metadata.AccountEmail),
				},
				store.UpsertSettingParams{
					Key:   KeyConnectedAt,
					Value: connectedAt.UTC().Format(time.RFC3339),
				},
			)
		}
		for _, setting := range settings {
			if err := q.UpsertSetting(ctx, setting); err != nil {
				return fmt.Errorf("save %s: %w", setting.Key, err)
			}
		}
		if tok.Expiry.IsZero() {
			if err := q.DeleteSetting(ctx, KeyTokenExpiry); err != nil {
				return fmt.Errorf("clear %s: %w", KeyTokenExpiry, err)
			}
		}
		if metadata != nil {
			if err := q.DeleteSetting(ctx, KeyFolderID); err != nil {
				return fmt.Errorf("clear %s: %w", KeyFolderID, err)
			}
			return nil
		}
		connectedAt, err := getSetting(ctx, q, KeyConnectedAt)
		if err != nil {
			return err
		}
		if connectedAt == "" {
			if err := q.UpsertSetting(ctx, store.UpsertSettingParams{
				Key:   KeyConnectedAt,
				Value: time.Now().UTC().Format(time.RFC3339),
			}); err != nil {
				return fmt.Errorf("save %s: %w", KeyConnectedAt, err)
			}
		}
		return nil
	})
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
	if s == nil || s.q == nil {
		return fmt.Errorf("settings store is not configured")
	}
	return store.InTransaction(ctx, s.q, func(q *store.Queries) error {
		for _, key := range allTokenKeys {
			if err := q.DeleteSetting(ctx, key); err != nil {
				return err
			}
		}
		return nil
	})
}
