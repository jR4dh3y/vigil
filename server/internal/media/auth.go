package media

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// AuthRequest is the JSON body MediaMTX POSTs to external auth (authMethod: http).
// See https://mediamtx.org/docs/features/authentication
type AuthRequest struct {
	User      string `json:"user"`
	Password  string `json:"password"`
	Token     string `json:"token"`
	IP        string `json:"ip"`
	Action    string `json:"action"`
	Path      string `json:"path"`
	Protocol  string `json:"protocol"`
	ID        string `json:"id"`
	Query     string `json:"query"`
	UserAgent string `json:"userAgent"`
}

// AuthHandler returns an HTTP handler for POST /internal/mediamtx/auth.
// Returns 200 if the stream token is valid for the path; 401 otherwise.
// This endpoint is internal (MediaMTX → Go); it is not part of the public OpenAPI surface.
func (s *Service) AuthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()

		var req AuthRequest
		dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
		if err := dec.Decode(&req); err != nil {
			// Also accept empty body as unauthenticated (MediaMTX probe for credentials).
			slog.Debug("mediamtx auth decode", "err", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if s.ValidateAuth(req) {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}
}
