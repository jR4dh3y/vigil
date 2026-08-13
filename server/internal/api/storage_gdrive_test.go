package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/recording"
	"github.com/nvr/nvr/server/internal/secrets"
	"github.com/nvr/nvr/server/internal/storage/gdrive"
	"github.com/nvr/nvr/server/internal/store"
)

const gDriveAPITestSecretsKey = "gdrive-api-test-secrets-key"

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
	server := newGDriveAPITestServer(t)
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
	if strings.Contains(response.Body.String(), "admin@example.com") {
		t.Fatalf("viewer response exposed account email: %s", response.Body.String())
	}
}

func TestGetGDriveStatusIncludesAccountEmailForAdmin(t *testing.T) {
	server := newGDriveAPITestServer(t)
	request := requestAsRole(http.MethodGet, "/api/v1/storage/gdrive/status", nil, auth.RoleAdmin)
	response := httptest.NewRecorder()

	server.GetGDriveStatus(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "admin@example.com") {
		t.Fatalf("admin response omitted account email: %s", response.Body.String())
	}
}

func TestGDriveMutationEndpointsRequireAdmin(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		call   func(*Server, http.ResponseWriter, *http.Request)
	}{
		{
			name:   "configuration",
			method: http.MethodPut,
			path:   "/api/v1/storage/gdrive/configuration",
			call:   (*Server).PutGDriveConfiguration,
		},
		{
			name:   "connect",
			method: http.MethodPost,
			path:   "/api/v1/storage/gdrive/connect",
			call:   (*Server).PostGDriveConnect,
		},
		{
			name:   "disconnect",
			method: http.MethodDelete,
			path:   "/api/v1/storage/gdrive/disconnect",
			call:   (*Server).DeleteGDriveDisconnect,
		},
		{
			name:   "archive",
			method: http.MethodPost,
			path:   "/api/v1/storage/gdrive/archive",
			call:   (*Server).PostGDriveArchive,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := &Server{}
			request := requestAsRole(test.method, test.path, nil, auth.RoleViewer)
			response := httptest.NewRecorder()

			test.call(server, response, request)

			if response.Code != http.StatusForbidden {
				t.Fatalf("got status %d, want %d: %s", response.Code, http.StatusForbidden, response.Body.String())
			}
		})
	}
}

func TestPutGDriveConfigurationStoresEncryptedCredentials(t *testing.T) {
	server := newGDriveAPITestServer(t)
	request := requestAsRole(
		http.MethodPut,
		"/api/v1/storage/gdrive/configuration",
		strings.NewReader(`{"clientId":"dashboard-client","clientSecret":"dashboard-secret","redirectUrl":"https://nvr.example.test/api/v1/storage/gdrive/callback"}`),
		auth.RoleAdmin,
	)
	response := httptest.NewRecorder()

	server.PutGDriveConfiguration(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"configured":true`) {
		t.Fatalf("configuration response missing configured state: %s", response.Body.String())
	}
	stored, err := server.Queries.GetSetting(context.Background(), gdrive.KeyOAuthSecret)
	if err != nil {
		t.Fatalf("read stored secret: %v", err)
	}
	if strings.Contains(stored.Value, "dashboard-secret") {
		t.Fatalf("OAuth client secret was stored in plaintext: %q", stored.Value)
	}
	if _, err := server.Queries.GetSetting(context.Background(), gdrive.KeyRefreshToken); err == nil {
		t.Fatal("old Drive refresh token was not cleared")
	}
}

func TestPutGDriveConfigurationRequiresClientSecret(t *testing.T) {
	server := newGDriveAPITestServer(t)
	request := requestAsRole(
		http.MethodPut,
		"/api/v1/storage/gdrive/configuration",
		strings.NewReader(`{"clientId":"dashboard-client","redirectUrl":"https://nvr.example.test/api/v1/storage/gdrive/callback"}`),
		auth.RoleAdmin,
	)
	response := httptest.NewRecorder()

	server.PutGDriveConfiguration(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"code":"validation"`) {
		t.Fatalf("unexpected validation response: %s", response.Body.String())
	}
}

func TestPostGDriveArchiveValidatesLimit(t *testing.T) {
	for _, limit := range []string{"0", "51"} {
		t.Run(limit, func(t *testing.T) {
			server := newGDriveAPITestServer(t)
			body := strings.NewReader(`{"limit":` + limit + `}`)
			request := requestAsRole(
				http.MethodPost,
				"/api/v1/storage/gdrive/archive",
				body,
				auth.RoleAdmin,
			)
			response := httptest.NewRecorder()

			server.PostGDriveArchive(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("got status %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), `"code":"validation"`) {
				t.Fatalf("unexpected validation response: %s", response.Body.String())
			}
		})
	}
}

func TestPostGDriveArchiveReturnsConflictWhenArchiveIsRunning(t *testing.T) {
	server := newGDriveAPITestServer(t)
	index := &blockingArchiveIndex{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	done := make(chan error, 1)
	go func() {
		_, err := server.GDrive.ArchivePending(context.Background(), index, 1)
		done <- err
	}()
	<-index.started

	request := requestAsRole(
		http.MethodPost,
		"/api/v1/storage/gdrive/archive",
		http.NoBody,
		auth.RoleAdmin,
	)
	response := httptest.NewRecorder()
	server.PostGDriveArchive(response, request)
	close(index.release)

	if err := <-done; err != nil {
		t.Fatalf("holding archive failed: %v", err)
	}
	if response.Code != http.StatusConflict {
		t.Fatalf("got status %d, want %d: %s", response.Code, http.StatusConflict, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"code":"archive_in_progress"`) {
		t.Fatalf("unexpected conflict response: %s", response.Body.String())
	}
}

type blockingArchiveIndex struct {
	started chan struct{}
	release chan struct{}
}

func (i *blockingArchiveIndex) ListUnarchived(context.Context, int) ([]gdrive.ArchiveSegment, error) {
	close(i.started)
	<-i.release
	return nil, nil
}

func (*blockingArchiveIndex) AbsolutePath(string) (string, error) {
	return "", nil
}

func (*blockingArchiveIndex) MarkArchived(context.Context, string, string) error {
	return nil
}

func (*blockingArchiveIndex) DeleteLocal(context.Context, string, string, string) error {
	return nil
}

func newGDriveAPITestServer(t *testing.T) *Server {
	t.Helper()
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate store: %v", err)
	}
	queries := store.New(db)
	accessToken, err := secrets.Encrypt(gDriveAPITestSecretsKey, "access-token")
	if err != nil {
		t.Fatalf("encrypt access token: %v", err)
	}
	refreshToken, err := secrets.Encrypt(gDriveAPITestSecretsKey, "refresh-token")
	if err != nil {
		t.Fatalf("encrypt refresh token: %v", err)
	}
	for _, setting := range []store.UpsertSettingParams{
		{Key: gdrive.KeyAccessToken, Value: accessToken},
		{Key: gdrive.KeyRefreshToken, Value: refreshToken},
		{Key: gdrive.KeyAccountEmail, Value: "admin@example.com"},
	} {
		if err := queries.UpsertSetting(context.Background(), setting); err != nil {
			t.Fatalf("save %s: %v", setting.Key, err)
		}
	}
	return &Server{
		Queries: queries,
		GDrive: gdrive.NewService(gdrive.Config{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RedirectURL:  "http://localhost/api/v1/storage/gdrive/callback",
		}, queries, gDriveAPITestSecretsKey),
		Recording: recording.NewService(queries, recording.Config{
			RecordingsDir: t.TempDir(),
			RetentionDays: 7,
		}),
	}
}

func requestAsRole(
	method string,
	path string,
	body io.Reader,
	role string,
) *http.Request {
	request := httptest.NewRequest(method, path, body)
	return request.WithContext(auth.WithUser(request.Context(), &auth.User{
		ID:   role + "-1",
		Role: role,
	}))
}
