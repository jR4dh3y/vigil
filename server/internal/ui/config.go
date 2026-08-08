package ui

import (
	"net/http"
	"net/url"
	"strings"
)

// Config carries the resolved public and hosted dashboard URLs used by the slim
// connection page. Full (embedded SPA) builds ignore it.
type Config struct {
	PublicURL          string
	HostedDashboardURL string
}

// BuildDashboardURL returns the hosted dashboard URL with its server query
// parameter set to the resolved public URL, falling back to the request origin.
// It returns an empty string when no hosted dashboard URL is configured or it
// cannot be parsed.
func BuildDashboardURL(hostedURL, publicURL string, r *http.Request) string {
	hosted := strings.TrimSpace(hostedURL)
	if hosted == "" {
		return ""
	}
	serverURL := strings.TrimSpace(publicURL)
	if serverURL == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		serverURL = scheme + "://" + r.Host
	}
	u, err := url.Parse(hosted)
	if err != nil {
		return ""
	}
	q := u.Query()
	q.Set("server", serverURL)
	u.RawQuery = q.Encode()
	return u.String()
}
