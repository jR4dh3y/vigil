package api

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/store"
)

// listUsersResponse matches GET /users OpenAPI response body.
type listUsersResponse struct {
	Users []UserPublic `json:"users"`
}

// ListUsers returns all users (admin only).
func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	rows, err := s.Queries.ListUsers(r.Context())
	if err != nil {
		slog.Error("list users", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	out := make([]UserPublic, 0, len(rows))
	for _, u := range rows {
		out = append(out, UserPublic{
			Id:       u.ID,
			Username: u.Username,
			Role:     UserPublicRole(u.Role),
		})
	}
	writeJSON(w, http.StatusOK, listUsersResponse{Users: out})
}

// CreateUser creates a user (admin only).
func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	var body CreateUserRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	username := strings.TrimSpace(body.Username)
	if username == "" {
		writeError(w, http.StatusBadRequest, "username is required", "validation")
		return
	}
	if utf8.RuneCountInString(body.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters", "validation")
		return
	}
	role := string(body.Role)
	switch role {
	case auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer:
	default:
		writeError(w, http.StatusBadRequest, "invalid role", "validation")
		return
	}

	// Unique username check before insert for a clean 409.
	if _, err := s.Queries.GetUserByUsername(r.Context(), username); err == nil {
		writeError(w, http.StatusConflict, "username already exists", "conflict")
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		slog.Error("get user by username", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		slog.Error("hash password", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	user, err := s.Queries.CreateUser(r.Context(), store.CreateUserParams{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: hash,
		Role:         role,
	})
	if err != nil {
		// Unique constraint race.
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "username already exists", "conflict")
			return
		}
		slog.Error("create user", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	writeJSON(w, http.StatusCreated, UserPublic{
		Id:       user.ID,
		Username: user.Username,
		Role:     UserPublicRole(user.Role),
	})
}

// DeleteUser deletes a user (admin only). Blocks self-delete and last-admin delete.
func (s *Server) DeleteUser(w http.ResponseWriter, r *http.Request, id string) {
	caller := requireAdmin(w, r)
	if caller == nil {
		return
	}

	id = strings.TrimSpace(id)
	if id == "" {
		writeError(w, http.StatusBadRequest, "user id is required", "validation")
		return
	}
	if id == caller.ID {
		writeError(w, http.StatusConflict, "cannot delete yourself", "conflict")
		return
	}

	target, err := s.Queries.GetUserByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "user not found", "not_found")
			return
		}
		slog.Error("get user", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	if target.Role == auth.RoleAdmin {
		admins, err := s.Queries.CountAdmins(r.Context())
		if err != nil {
			slog.Error("count admins", "err", err)
			writeError(w, http.StatusInternalServerError, "internal error", "internal")
			return
		}
		if admins <= 1 {
			writeError(w, http.StatusConflict, "cannot delete the last admin", "conflict")
			return
		}
	}

	if err := s.Queries.DeleteUser(r.Context(), id); err != nil {
		slog.Error("delete user", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// modernc.org/sqlite / database/sql surface UNIQUE via error string.
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "constraint")
}
