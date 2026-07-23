package auth

import (
	"net/http"
	"strings"
	"time"
)

// HeaderSessionToken is an alternate way for non-browser clients (e.g. React Native)
// to receive and send the opaque session token when cookie jars are unreliable.
const HeaderSessionToken = "X-Session-Token"

// SetSessionCookie writes the nvr_session cookie (HttpOnly, Path=/, SameSite=Lax).
// Secure is left false for local HTTP; reverse proxies terminate TLS in production.
func SetSessionCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})
}

// ClearSessionCookie expires the session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})
}

// WriteSessionTokenHeader exposes the raw token for mobile / API clients.
func WriteSessionTokenHeader(w http.ResponseWriter, token string) {
	w.Header().Set(HeaderSessionToken, token)
}

// SessionTokenFromRequest reads the session token from cookie, Authorization Bearer,
// or X-Session-Token (cookie wins when multiple are present).
func SessionTokenFromRequest(r *http.Request) (string, bool) {
	if c, err := r.Cookie(CookieName); err == nil && c.Value != "" {
		return c.Value, true
	}
	if authz := r.Header.Get("Authorization"); authz != "" {
		const prefix = "bearer "
		if len(authz) > len(prefix) && strings.EqualFold(authz[:len(prefix)], prefix) {
			token := strings.TrimSpace(authz[len(prefix):])
			if token != "" {
				return token, true
			}
		}
	}
	if token := strings.TrimSpace(r.Header.Get(HeaderSessionToken)); token != "" {
		return token, true
	}
	return "", false
}
