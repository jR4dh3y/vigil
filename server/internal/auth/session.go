package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	// SessionTTL is how long a session cookie remains valid.
	SessionTTL = 30 * 24 * time.Hour

	// CookieName is the HTTP cookie holding the opaque session token.
	CookieName = "nvr_session"
)

// Role constants matching the users.role CHECK constraint and OpenAPI enum.
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)

// NewSessionToken generates a cryptographically random opaque token and its SHA-256 hash.
// Store only the hash; return the raw token to the client in the cookie.
func NewSessionToken() (token string, tokenHash string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("generate session token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	tokenHash = HashToken(token)
	return token, tokenHash, nil
}

// HashToken returns the hex-encoded SHA-256 of token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// SessionExpiresAt returns the expiry time for a new session.
func SessionExpiresAt() time.Time {
	return time.Now().UTC().Add(SessionTTL)
}

// FormatSQLiteTime formats t for storage in SQLite TEXT datetime columns.
func FormatSQLiteTime(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04:05")
}

// ParseSQLiteTime parses a SQLite datetime('now') style timestamp.
func ParseSQLiteTime(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.UTC); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse time %q", s)
}
