package gdrive

import (
	"fmt"
	"net/url"
	"strings"
)

// Config holds Google OAuth client credentials for Drive connect.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	// APIEndpoint optionally overrides the Drive REST base URL for tests or an
	// API-compatible emulator. Production leaves it empty.
	APIEndpoint string
}

// Configured reports whether OAuth client credentials are fully set.
func (c Config) Configured() bool {
	return c.Validate() == nil
}

// Validate checks the OAuth settings needed to start a browser flow.
func (c Config) Validate() error {
	if strings.TrimSpace(c.ClientID) == "" ||
		strings.TrimSpace(c.ClientSecret) == "" ||
		strings.TrimSpace(c.RedirectURL) == "" {
		return fmt.Errorf("client id, client secret, and redirect URL are required")
	}
	redirect, err := url.Parse(c.RedirectURL)
	if err != nil || redirect.Host == "" || (redirect.Scheme != "http" && redirect.Scheme != "https") {
		return fmt.Errorf("redirect URL must be an absolute HTTP(S) URL")
	}
	return nil
}
