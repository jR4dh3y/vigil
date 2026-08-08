package api

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/nvr/nvr/server/internal/auth"
)

// GetHealth returns service health.
func (s *Server) GetHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Health{Status: HealthStatusOk})
}

// GetVersion returns build version metadata.
func (s *Server) GetVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Version{
		Version: s.Version,
		Commit:  s.Commit,
	})
}

// GetAuthStatus reports whether setup is required and the current session user.
func (s *Server) GetAuthStatus(w http.ResponseWriter, r *http.Request) {
	count, err := s.Queries.CountUsers(r.Context())
	if err != nil {
		slog.Error("count users", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	status := AuthStatus{SetupRequired: count == 0}
	if u := auth.UserFromContext(r.Context()); u != nil {
		pub := toUserPublic(u)
		status.User = &pub
	}
	writeJSON(w, http.StatusOK, status)
}

// PostAuthSetup creates the first admin user and starts a session.
func (s *Server) PostAuthSetup(w http.ResponseWriter, r *http.Request) {
	var body SetupRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	user, err := auth.CreateFirstAdmin(r.Context(), s.Queries, body.Username, body.Password)
	switch {
	case errors.Is(err, auth.ErrInvalidAdminCredentials):
		writeError(w, http.StatusBadRequest, err.Error(), "validation")
		return
	case errors.Is(err, auth.ErrSetupComplete):
		writeError(w, http.StatusConflict, "setup already completed", "setup_complete")
		return
	case err != nil:
		slog.Error("create first admin", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	if err := s.issueSession(w, r, user.ID); err != nil {
		slog.Error("create session", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	writeJSON(w, http.StatusCreated, UserPublic{
		Id:       user.ID,
		Username: user.Username,
		Role:     UserPublicRole(user.Role),
	})
}

// PostAuthLogin authenticates and sets a session cookie.
func (s *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) {
	var body LoginRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	username := strings.TrimSpace(body.Username)
	if username == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password required", "validation")
		return
	}

	user, err := s.Queries.GetUserByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "invalid credentials", "unauthorized")
			return
		}
		slog.Error("get user", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	ok, err := auth.VerifyPassword(body.Password, user.PasswordHash)
	if err != nil {
		slog.Error("verify password", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	if !ok {
		writeError(w, http.StatusUnauthorized, "invalid credentials", "unauthorized")
		return
	}

	if err := s.issueSession(w, r, user.ID); err != nil {
		slog.Error("create session", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	writeJSON(w, http.StatusOK, UserPublic{
		Id:       user.ID,
		Username: user.Username,
		Role:     UserPublicRole(user.Role),
	})
}

// PostAuthLogout clears the session cookie and deletes the session row.
func (s *Server) PostAuthLogout(w http.ResponseWriter, r *http.Request) {
	if token, ok := auth.SessionTokenFromRequest(r); ok {
		if err := s.Auth.DeleteSession(r.Context(), token); err != nil {
			slog.Warn("delete session", "err", err)
		}
	}
	auth.ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// GetAuthMe returns the authenticated user.
func (s *Server) GetAuthMe(w http.ResponseWriter, r *http.Request) {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, toUserPublic(u))
}

func (s *Server) issueSession(w http.ResponseWriter, r *http.Request, userID string) error {
	token, expires, err := s.Auth.CreateSession(r.Context(), userID)
	if err != nil {
		return err
	}
	auth.SetSessionCookie(w, token, expires)
	// Mobile clients (React Native) often cannot read Set-Cookie; mirror the token.
	auth.WriteSessionTokenHeader(w, token)
	return nil
}

func toUserPublic(u *auth.User) UserPublic {
	return UserPublic{
		Id:       u.ID,
		Username: u.Username,
		Role:     UserPublicRole(u.Role),
	}
}
