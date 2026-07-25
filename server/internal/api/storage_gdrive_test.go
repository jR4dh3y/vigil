package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nvr/nvr/server/internal/auth"
)

func TestGetGDriveStatusRequiresAuthentication(t *testing.T) {
	server := &Server{}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/storage/gdrive/status", nil)
	response := httptest.NewRecorder()

	server.GetGDriveStatus(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestGetGDriveStatusAllowsNonAdminUsers(t *testing.T) {
	server := &Server{}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/storage/gdrive/status", nil)
	request = request.WithContext(auth.WithUser(request.Context(), &auth.User{
		ID:   "viewer-1",
		Role: auth.RoleViewer,
	}))
	response := httptest.NewRecorder()

	server.GetGDriveStatus(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
}
