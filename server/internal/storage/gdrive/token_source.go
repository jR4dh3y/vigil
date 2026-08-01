package gdrive

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"golang.org/x/oauth2"
)

// persistingTokenSource wraps an oauth2.TokenSource and re-saves tokens when
// the access token is refreshed, preserving the refresh token (which Google
// often omits from refresh responses).
type persistingTokenSource struct {
	src     oauth2.TokenSource
	svc     *Service
	mu      sync.Mutex
	last    *oauth2.Token
	refresh string
}

func (s *Service) tokenSource(ctx context.Context, cfg Config) (oauth2.TokenSource, error) {
	tok, err := s.loadToken(ctx)
	if err != nil {
		return nil, err
	}
	if tok == nil || tok.RefreshToken == "" {
		return nil, fmt.Errorf("google drive is not connected")
	}
	base := oauthConfig(cfg).TokenSource(ctx, tok)
	return &persistingTokenSource{
		src:     base,
		svc:     s,
		last:    tok,
		refresh: tok.RefreshToken,
	}, nil
}

func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	tok, err := p.src.Token()
	if err != nil {
		return nil, err
	}
	if tok == nil {
		return nil, fmt.Errorf("nil token from source")
	}

	// Always preserve refresh token for subsequent saves / callers.
	if tok.RefreshToken == "" {
		tok.RefreshToken = p.refresh
	} else {
		p.refresh = tok.RefreshToken
	}

	// Re-persist when access token changes (refresh occurred).
	if p.last == nil || tok.AccessToken != p.last.AccessToken {
		toSave := *tok
		if toSave.RefreshToken == "" {
			toSave.RefreshToken = p.refresh
		}
		if err := p.svc.saveToken(context.Background(), &toSave, nil); err != nil {
			slog.Warn("gdrive persist refreshed token", "err", err)
		}
		p.last = tok
	}
	return tok, nil
}
