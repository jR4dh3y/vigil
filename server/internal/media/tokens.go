package media

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// DefaultTokenTTL is the lifetime of stream tokens issued for live view.
// Tokens are reusable within TTL (needed for HLS playlists/segments); not strict one-time.
const DefaultTokenTTL = 60 * time.Second

// TokenEntry is a minted stream token record.
type TokenEntry struct {
	CameraID  string
	Path      string
	ExpiresAt time.Time
}

// TokenStore is an in-memory map of stream tokens.
// Tokens are short-lived (~60s) and reusable until expiry so HLS segment fetches succeed.
type TokenStore struct {
	mu     sync.Mutex
	tokens map[string]TokenEntry
}

// NewTokenStore creates an empty token store.
func NewTokenStore() *TokenStore {
	return &TokenStore{tokens: make(map[string]TokenEntry)}
}

// MintToken creates a new random token for cameraID/path with the given TTL.
// If ttl <= 0, DefaultTokenTTL is used.
func (s *TokenStore) MintToken(cameraID, path string, ttl time.Duration) (token string, expiresAt time.Time, err error) {
	if ttl <= 0 {
		ttl = DefaultTokenTTL
	}
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}
	token = hex.EncodeToString(raw)
	expiresAt = time.Now().Add(ttl)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked(time.Now())
	s.tokens[token] = TokenEntry{
		CameraID:  cameraID,
		Path:      path,
		ExpiresAt: expiresAt,
	}
	return token, expiresAt, nil
}

// ValidateAndConsume checks that token is valid for path (or any path if path is empty).
// Tokens are reusable within TTL (not deleted on success) so HLS playlists can re-auth.
// Expired entries are purged. Returns false if missing, expired, or path mismatch.
func (s *TokenStore) ValidateAndConsume(token, path string) bool {
	if token == "" {
		return false
	}
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked(now)

	entry, ok := s.tokens[token]
	if !ok {
		return false
	}
	if now.After(entry.ExpiresAt) {
		delete(s.tokens, token)
		return false
	}
	if path != "" && entry.Path != path {
		return false
	}
	return true
}

// Len returns the number of tokens currently stored (for tests).
func (s *TokenStore) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked(time.Now())
	return len(s.tokens)
}

func (s *TokenStore) purgeLocked(now time.Time) {
	for k, v := range s.tokens {
		if now.After(v.ExpiresAt) {
			delete(s.tokens, k)
		}
	}
}
