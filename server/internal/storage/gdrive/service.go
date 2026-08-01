package gdrive

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/nvr/nvr/server/internal/secrets"
	"github.com/nvr/nvr/server/internal/store"
	"golang.org/x/oauth2"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

const (
	oauthStateTTL   = 10 * time.Minute
	oauthStateBytes = 32
	// Frontend settings path for post-callback redirects.
	settingsRedirectPath = "/settings"
)

// Service manages Google Drive OAuth connection lifecycle and archive uploads.
type Service struct {
	cfg        Config
	q          *store.Queries
	secretsKey string
	httpClient *http.Client
	// archiveMu serializes ArchivePending (cron + API) to avoid double uploads.
	archiveMu sync.Mutex
	// archiveFile is a test seam for the batch orchestration. Production calls
	// ArchiveLocalFile directly.
	archiveFile func(context.Context, string, string, string) (string, error)
}

// NewService constructs a gdrive Service.
func NewService(cfg Config, q *store.Queries, secretsKey string) *Service {
	return &Service{
		cfg:        cfg,
		q:          q,
		secretsKey: secretsKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// Status is a public view of Drive connection (never includes tokens).
type Status struct {
	Configured      bool
	Connected       bool
	ConnectionError string
	AccountEmail    string
	ConnectedAt     string // RFC3339 if set
	FolderID        string
}

// Configure saves OAuth client credentials supplied by an administrator. The
// secret is encrypted at rest, and existing account tokens are cleared because
// they belong to the previous OAuth client configuration.
func (s *Service) Configure(ctx context.Context, cfg Config) error {
	if s == nil || s.q == nil {
		return fmt.Errorf("settings store is not configured")
	}
	cfg.ClientID = strings.TrimSpace(cfg.ClientID)
	cfg.ClientSecret = strings.TrimSpace(cfg.ClientSecret)
	cfg.RedirectURL = strings.TrimSpace(cfg.RedirectURL)
	if err := cfg.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(s.secretsKey) == "" {
		return fmt.Errorf("NVR_SECRETS_KEY is required to store Google Drive configuration securely")
	}

	secret, err := secrets.Encrypt(s.secretsKey, cfg.ClientSecret)
	if err != nil {
		return fmt.Errorf("encrypt OAuth client secret: %w", err)
	}
	return store.InTransaction(ctx, s.q, func(q *store.Queries) error {
		for _, setting := range []store.UpsertSettingParams{
			{Key: KeyOAuthClientID, Value: cfg.ClientID},
			{Key: KeyOAuthSecret, Value: secret},
			{Key: KeyOAuthRedirect, Value: cfg.RedirectURL},
		} {
			if err := q.UpsertSetting(ctx, setting); err != nil {
				return fmt.Errorf("save %s: %w", setting.Key, err)
			}
		}
		for _, key := range allTokenKeys {
			if err := q.DeleteSetting(ctx, key); err != nil {
				return fmt.Errorf("clear %s: %w", key, err)
			}
		}
		return nil
	})
}

// Status returns whether OAuth is configured and whether an account is connected.
func (s *Service) Status(ctx context.Context) (Status, error) {
	cfg, err := s.currentConfig(ctx)
	if err != nil {
		return Status{ConnectionError: "Saved Google Drive configuration cannot be decrypted. Save it again before reconnecting."}, nil
	}
	// Configured means the admin can start connect: OAuth app config + secrets key for token encryption.
	st := Status{Configured: cfg.Configured() && strings.TrimSpace(s.secretsKey) != ""}
	if s.q == nil {
		return st, nil
	}

	refresh, err := s.getSetting(ctx, KeyRefreshToken)
	if err != nil {
		return st, err
	}
	if refresh != "" {
		if !st.Configured {
			st.ConnectionError = "Google Drive is linked, but the server OAuth or secrets configuration is incomplete."
		} else if _, err := s.loadToken(ctx); err != nil {
			// Keep status usable so an administrator can disconnect or reconnect
			// after rotating/misconfiguring NVR_SECRETS_KEY.
			slog.Warn("gdrive stored credentials are unusable", "err", err)
			st.ConnectionError = "Stored Google Drive credentials cannot be decrypted. Reconnect Google Drive."
		} else {
			st.Connected = true
		}
	}

	if email, err := s.getSetting(ctx, KeyAccountEmail); err != nil {
		return st, err
	} else {
		st.AccountEmail = email
	}
	if at, err := s.getSetting(ctx, KeyConnectedAt); err != nil {
		return st, err
	} else {
		st.ConnectedAt = at
	}
	if folder, err := s.getSetting(ctx, KeyFolderID); err != nil {
		return st, err
	} else {
		st.FolderID = folder
	}
	return st, nil
}

// BeginConnect starts the OAuth flow and returns the Google authorization URL.
func (s *Service) BeginConnect(ctx context.Context) (authorizationURL string, err error) {
	cfg, err := s.currentConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("load google drive oauth configuration: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return "", fmt.Errorf("google drive oauth is not configured: %w", err)
	}
	if strings.TrimSpace(s.secretsKey) == "" {
		return "", fmt.Errorf("NVR_SECRETS_KEY is required to store Google Drive tokens securely")
	}
	state, err := randomState()
	if err != nil {
		return "", err
	}
	expires := time.Now().UTC().Add(oauthStateTTL)
	if err := s.saveOAuthState(ctx, state, expires); err != nil {
		return "", fmt.Errorf("save oauth state: %w", err)
	}
	return s.authCodeURL(cfg, state), nil
}

// CallbackResult describes the outcome of HandleCallback for HTTP redirects.
type CallbackResult struct {
	OK      bool
	Message string
}

// HandleCallback exchanges the OAuth code, stores encrypted tokens, and fetches account email.
// On provider error query params, returns a failed result without exchanging.
// CSRF state is always required (including Google error responses) so unauthenticated
// callers cannot inject arbitrary flash messages into the settings UI.
func (s *Service) HandleCallback(ctx context.Context, code, state, errParam, errDesc string) (CallbackResult, error) {
	if state == "" {
		return CallbackResult{OK: false, Message: "missing state"}, nil
	}
	if err := s.consumeOAuthState(ctx, state); err != nil {
		return CallbackResult{OK: false, Message: "invalid or expired state"}, nil
	}

	if errParam != "" {
		// Do not reflect raw Google error_description into the UI (phishing risk).
		slog.Info("gdrive oauth denied or failed", "error", errParam, "description", errDesc)
		msg := "access denied"
		if errParam != "access_denied" {
			msg = "authorization failed"
		}
		return CallbackResult{OK: false, Message: msg}, nil
	}
	if code == "" {
		return CallbackResult{OK: false, Message: "missing code"}, nil
	}
	cfg, err := s.currentConfig(ctx)
	if err != nil || !cfg.Configured() {
		return CallbackResult{OK: false, Message: "google drive oauth is not configured"}, nil
	}
	if strings.TrimSpace(s.secretsKey) == "" {
		return CallbackResult{OK: false, Message: "server secrets key is not configured"}, nil
	}

	tok, err := s.exchange(ctx, cfg, code)
	if err != nil {
		slog.Error("gdrive oauth exchange", "err", err)
		return CallbackResult{OK: false, Message: "token exchange failed"}, nil
	}
	if tok.RefreshToken == "" {
		return CallbackResult{OK: false, Message: "no refresh token returned; disconnect and reconnect with consent"}, nil
	}

	email, err := s.fetchAccountEmail(ctx, cfg, tok)
	if err != nil {
		slog.Warn("gdrive fetch account email", "err", err)
		// Non-fatal: still store tokens.
		email = ""
	}

	if err := s.saveToken(ctx, tok, &connectionMetadata{
		AccountEmail: email,
		ConnectedAt:  time.Now().UTC(),
	}); err != nil {
		return CallbackResult{}, fmt.Errorf("save token: %w", err)
	}
	return CallbackResult{OK: true, Message: "connected"}, nil
}

// Disconnect clears stored gdrive settings and best-effort revokes the token.
func (s *Service) Disconnect(ctx context.Context) error {
	tok, err := s.loadToken(ctx)
	if err != nil {
		slog.Warn("gdrive load token for revoke", "err", err)
	} else if tok != nil {
		s.revokeBestEffort(ctx, tok)
	}
	return s.clearAllGDriveSettings(ctx)
}

// RedirectURL builds the frontend redirect after OAuth callback.
func RedirectURL(result CallbackResult) string {
	q := url.Values{}
	if result.OK {
		q.Set("gdrive", "connected")
	} else {
		q.Set("gdrive", "error")
		if result.Message != "" {
			q.Set("message", result.Message)
		}
	}
	return settingsRedirectPath + "?" + q.Encode()
}

func (s *Service) fetchAccountEmail(ctx context.Context, cfg Config, tok *oauth2.Token) (string, error) {
	client := oauthConfig(cfg).Client(ctx, tok)
	svc, err := oauth2api.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", err
	}
	ui, err := svc.Userinfo.Get().Do()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(ui.Email), nil
}

func (s *Service) currentConfig(ctx context.Context) (Config, error) {
	if s == nil {
		return Config{}, fmt.Errorf("google drive service is nil")
	}
	if s.q == nil {
		return s.cfg, nil
	}

	clientID, err := s.getSetting(ctx, KeyOAuthClientID)
	if err != nil {
		return Config{}, err
	}
	secret, err := s.getSetting(ctx, KeyOAuthSecret)
	if err != nil {
		return Config{}, err
	}
	redirectURL, err := s.getSetting(ctx, KeyOAuthRedirect)
	if err != nil {
		return Config{}, err
	}
	if clientID == "" && secret == "" && redirectURL == "" {
		return s.cfg, nil
	}
	decryptedSecret, err := secrets.Decrypt(s.secretsKey, secret)
	if err != nil {
		return Config{}, fmt.Errorf("decrypt OAuth client secret: %w", err)
	}
	return Config{
		ClientID:     clientID,
		ClientSecret: decryptedSecret,
		RedirectURL:  redirectURL,
		APIEndpoint:  s.cfg.APIEndpoint,
	}, nil
}

func (s *Service) revokeBestEffort(ctx context.Context, tok *oauth2.Token) {
	token := tok.RefreshToken
	if token == "" {
		token = tok.AccessToken
	}
	if token == "" {
		return
	}
	form := url.Values{"token": {token}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/revoke", strings.NewReader(form.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.Warn("gdrive token revoke", "err", err)
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		slog.Warn("gdrive token revoke status", "status", resp.StatusCode)
	}
}

func randomState() (string, error) {
	b := make([]byte, oauthStateBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}
	return hex.EncodeToString(b), nil
}
