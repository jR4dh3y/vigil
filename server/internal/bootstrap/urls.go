// Package bootstrap holds first-run setup helpers shared by the CLI, the HTTP
// setup handler, and server startup: admin creation, URL resolution/validation,
// and CORS origin derivation.
package bootstrap

import (
	"fmt"
	"net/url"
	"strings"
)

// Setting keys persisted in the settings KV table.
const (
	SettingPublicURL          = "publicUrl"
	SettingHostedDashboardURL = "hostedDashboardUrl"
)

// ValidateURL checks that s is a non-empty absolute http(s) URL.
func ValidateURL(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("URL is empty")
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid URL %q", s)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL %q must be http or https", s)
	}
	return nil
}

// Origin returns the scheme://host origin of a validated http(s) URL.
func Origin(rawURL string) (string, error) {
	if err := ValidateURL(rawURL); err != nil {
		return "", err
	}
	u, _ := url.Parse(strings.TrimSpace(rawURL))
	return u.Scheme + "://" + u.Host, nil
}

// ResolveURL returns env when non-empty, otherwise db. It implements the
// env-over-DB precedence for public and hosted dashboard URLs.
func ResolveURL(env, db string) string {
	if v := strings.TrimSpace(env); v != "" {
		return v
	}
	return strings.TrimSpace(db)
}

// IsLocalOrigin reports whether origin is a localhost development origin.
func IsLocalOrigin(origin string) bool {
	return strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")
}

// CORSOrigins builds the exact allowed origin list from configured env origins
// plus the hosted dashboard origin. Non-local configured origins must be HTTPS.
func CORSOrigins(envOrigins []string, hostedDashboardURL string) ([]string, error) {
	var origins []string
	seen := map[string]bool{}
	for _, o := range envOrigins {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		orig, err := Origin(o)
		if err != nil {
			return nil, err
		}
		if !IsLocalOrigin(orig) && !strings.HasPrefix(orig, "https://") {
			return nil, fmt.Errorf("CORS origin %q must be HTTPS", o)
		}
		if !seen[orig] {
			seen[orig] = true
			origins = append(origins, orig)
		}
	}
	if strings.TrimSpace(hostedDashboardURL) != "" {
		orig, err := Origin(hostedDashboardURL)
		if err != nil {
			return nil, fmt.Errorf("hosted dashboard URL: %w", err)
		}
		if !seen[orig] {
			seen[orig] = true
			origins = append(origins, orig)
		}
	}
	return origins, nil
}
