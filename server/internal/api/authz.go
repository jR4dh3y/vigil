package api

import (
	"net/http"

	"github.com/nvr/nvr/server/internal/auth"
)

// requireUser ensures the caller is authenticated.
// Returns the user, or nil after writing an error response.
func requireUser(w http.ResponseWriter, r *http.Request) *auth.User {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "unauthorized")
		return nil
	}
	return u
}

// requireAdmin ensures the caller is authenticated as admin.
// Returns the user, or nil after writing an error response.
func requireAdmin(w http.ResponseWriter, r *http.Request) *auth.User {
	u := requireUser(w, r)
	if u == nil {
		return nil
	}
	if u.Role != auth.RoleAdmin {
		writeError(w, http.StatusForbidden, "admin role required", "forbidden")
		return nil
	}
	return u
}

// requireOperator ensures the caller is authenticated as admin or operator.
// Returns false after writing an error response.
func requireOperator(w http.ResponseWriter, r *http.Request) bool {
	u := auth.UserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "unauthorized")
		return false
	}
	if u.Role != auth.RoleAdmin && u.Role != auth.RoleOperator {
		writeError(w, http.StatusForbidden, "insufficient role", "forbidden")
		return false
	}
	return true
}
