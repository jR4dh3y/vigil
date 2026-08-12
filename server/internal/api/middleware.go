package api

import (
	"log/slog"
	"net/http"

	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/bootstrap"
)

// SessionMiddleware loads the session cookie (if any) and attaches the user to the request context.
func SessionMiddleware(authSvc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token, ok := auth.SessionTokenFromRequest(r); ok {
				user, err := authSvc.UserFromToken(r.Context(), token)
				if err != nil {
					slog.Warn("session lookup failed", "err", err)
				} else if user != nil {
					r = r.WithContext(auth.WithUser(r.Context(), user))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware allows configured exact origins plus localhost development
// origins. Configured (hosted) origins receive allow-origin, methods, the
// Authorization and X-Session-Token allow-headers, the X-Session-Token expose
// header, and Vary: Origin, but no credential allowance. Localhost dev origins
// additionally retain Allow-Credentials. Unknown origins receive no grants.
func CORSMiddleware(configured []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (bootstrap.IsLocalOrigin(origin) || inOrigins(configured, origin)) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-Token, Range")
				w.Header().Set("Access-Control-Expose-Headers", "X-Session-Token, Accept-Ranges, Content-Length, Content-Range")
				w.Header().Set("Vary", "Origin")
				if bootstrap.IsLocalOrigin(origin) {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func inOrigins(origins []string, origin string) bool {
	for _, o := range origins {
		if o == origin {
			return true
		}
	}
	return false
}
