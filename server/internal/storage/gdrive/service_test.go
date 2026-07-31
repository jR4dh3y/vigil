package gdrive

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nvr/nvr/server/internal/store"
	"golang.org/x/oauth2"
)

const testSecretsKey = "test-drive-secrets-key"

func newTestService(t *testing.T) *Service {
	t.Helper()
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate store: %v", err)
	}
	service := NewService(Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost/api/v1/storage/gdrive/callback",
	}, store.New(db), testSecretsKey)
	if err := service.saveToken(context.Background(), &oauth2.Token{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}, &connectionMetadata{
		AccountEmail: "admin@example.com",
		ConnectedAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save token: %v", err)
	}
	return service
}

func TestStatusReportsUsableConnection(t *testing.T) {
	service := newTestService(t)
	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.Configured || !status.Connected {
		t.Fatalf("unexpected status: %+v", status)
	}
	if status.AccountEmail != "admin@example.com" || status.ConnectionError != "" {
		t.Fatalf("unexpected connection details: %+v", status)
	}
}

func TestConfigRequiresAbsoluteRedirectURL(t *testing.T) {
	config := Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "/api/v1/storage/gdrive/callback",
	}
	if config.Configured() {
		t.Fatal("relative redirect URL must not be considered configured")
	}
}

func TestConfigureStoresOAuthClientSecretEncrypted(t *testing.T) {
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate store: %v", err)
	}
	queries := store.New(db)
	service := NewService(Config{}, queries, testSecretsKey)
	config := Config{
		ClientID:     "dashboard-client-id",
		ClientSecret: "dashboard-client-secret",
		RedirectURL:  "https://nvr.example.test/api/v1/storage/gdrive/callback",
	}
	if err := service.Configure(context.Background(), config); err != nil {
		t.Fatalf("Configure: %v", err)
	}

	storedSecret, err := queries.GetSetting(context.Background(), KeyOAuthSecret)
	if err != nil {
		t.Fatalf("read stored secret: %v", err)
	}
	if strings.Contains(storedSecret.Value, config.ClientSecret) {
		t.Fatalf("OAuth client secret was stored in plaintext: %q", storedSecret.Value)
	}
	activeConfig, err := service.currentConfig(context.Background())
	if err != nil {
		t.Fatalf("currentConfig: %v", err)
	}
	if activeConfig != config {
		t.Fatalf("stored configuration mismatch: got %+v, want %+v", activeConfig, config)
	}
	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.Configured || status.Connected {
		t.Fatalf("unexpected status after configuration: %+v", status)
	}
}

func TestStatusAllowsRecoveryFromWrongSecretsKey(t *testing.T) {
	service := newTestService(t)
	service.secretsKey = "different-key"

	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Connected {
		t.Fatalf("expected unusable connection: %+v", status)
	}
	if status.ConnectionError == "" {
		t.Fatalf("expected recoverable connection error: %+v", status)
	}
}

func TestOAuthStateIsSingleUse(t *testing.T) {
	service := newTestService(t)
	authorizationURL, err := service.BeginConnect(context.Background())
	if err != nil {
		t.Fatalf("BeginConnect: %v", err)
	}
	parsed, err := url.Parse(authorizationURL)
	if err != nil {
		t.Fatalf("parse authorization URL: %v", err)
	}
	state := parsed.Query().Get("state")
	if state == "" {
		t.Fatal("authorization URL has no state")
	}

	result, err := service.HandleCallback(
		context.Background(),
		"",
		state,
		"access_denied",
		"provider-controlled text",
	)
	if err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}
	if result.OK || result.Message != "access denied" {
		t.Fatalf("unexpected denial result: %+v", result)
	}

	reused, err := service.HandleCallback(
		context.Background(),
		"",
		state,
		"access_denied",
		"",
	)
	if err != nil {
		t.Fatalf("HandleCallback reused state: %v", err)
	}
	if reused.OK || reused.Message != "invalid or expired state" {
		t.Fatalf("state was reusable: %+v", reused)
	}
}

func TestOAuthStateAllowsConcurrentConnectionFlows(t *testing.T) {
	service := newTestService(t)

	firstURL, err := service.BeginConnect(context.Background())
	if err != nil {
		t.Fatalf("first BeginConnect: %v", err)
	}
	secondURL, err := service.BeginConnect(context.Background())
	if err != nil {
		t.Fatalf("second BeginConnect: %v", err)
	}

	for name, authorizationURL := range map[string]string{
		"first":  firstURL,
		"second": secondURL,
	} {
		t.Run(name, func(t *testing.T) {
			parsed, err := url.Parse(authorizationURL)
			if err != nil {
				t.Fatalf("parse authorization URL: %v", err)
			}
			result, err := service.HandleCallback(
				context.Background(),
				"",
				parsed.Query().Get("state"),
				"access_denied",
				"",
			)
			if err != nil {
				t.Fatalf("HandleCallback: %v", err)
			}
			if result.OK || result.Message != "access denied" {
				t.Fatalf("unexpected callback result: %+v", result)
			}
		})
	}
}

func TestSaveConnectedTokenClearsStaleAccountMetadata(t *testing.T) {
	service := newTestService(t)
	if err := service.setSetting(context.Background(), KeyFolderID, "old-folder"); err != nil {
		t.Fatalf("set old folder: %v", err)
	}

	if err := service.saveToken(context.Background(), &oauth2.Token{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}, &connectionMetadata{
		AccountEmail: "",
		ConnectedAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save replacement token: %v", err)
	}

	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.AccountEmail != "" {
		t.Fatalf("stale account email retained: %q", status.AccountEmail)
	}
	if status.FolderID != "" {
		t.Fatalf("stale folder id retained: %q", status.FolderID)
	}
}

func TestEscapeDriveQuery(t *testing.T) {
	got := escapeDriveQuery(`folder\name's`)
	if got != `folder\\name\'s` {
		t.Fatalf("got %q", got)
	}
}

func TestUploadIsIdempotentByRecordingID(t *testing.T) {
	service := newTestService(t)
	var uploadCount int
	var uploaded bool
	driveAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/files/archive-folder":
			writeDriveJSON(t, w, `{"id":"archive-folder","mimeType":"application/vnd.google-apps.folder","trashed":false}`)
		case r.Method == http.MethodGet && r.URL.Path == "/files":
			query := r.URL.Query().Get("q")
			if strings.Contains(query, "appProperties") && uploaded {
				writeDriveJSON(t, w, `{"files":[{"id":"drive-file-1"}]}`)
				return
			}
			if strings.Contains(query, "name = 'NVR Archives'") {
				writeDriveJSON(t, w, `{"files":[{"id":"archive-folder","name":"NVR Archives"}]}`)
				return
			}
			writeDriveJSON(t, w, `{"files":[]}`)
		case r.Method == http.MethodPost && r.URL.Path == "/upload/drive/v3/files":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read upload body: %v", err)
			}
			if !strings.Contains(string(body), recordingIDProperty) ||
				!strings.Contains(string(body), "recording-1") {
				t.Errorf("upload metadata lacks recording id: %s", body)
			}
			uploadCount++
			uploaded = true
			writeDriveJSON(t, w, `{"id":"drive-file-1"}`)
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
			t.Errorf("unexpected Drive request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer driveAPI.Close()
	service.cfg.APIEndpoint = driveAPI.URL + "/"

	for range 2 {
		id, err := service.Upload(
			context.Background(),
			"recording-1",
			"camera/segment.mp4",
			strings.NewReader("recording"),
			int64(len("recording")),
			"video/mp4",
		)
		if err != nil {
			t.Fatalf("Upload: %v", err)
		}
		if id != "drive-file-1" {
			t.Fatalf("got file id %q", id)
		}
	}
	if uploadCount != 1 {
		t.Fatalf("uploaded %d times, want 1", uploadCount)
	}
}

type fakeArchiveIndex struct {
	segments  []ArchiveSegment
	paths     map[string]string
	marked    map[string]string
	markError error
}

func (f *fakeArchiveIndex) ListUnarchived(context.Context, int) ([]ArchiveSegment, error) {
	return f.segments, nil
}

func (f *fakeArchiveIndex) AbsolutePath(rel string) (string, error) {
	path, ok := f.paths[rel]
	if !ok {
		return "", errors.New("unknown path")
	}
	return path, nil
}

func (f *fakeArchiveIndex) MarkArchived(_ context.Context, id, location string) error {
	if f.markError != nil {
		return f.markError
	}
	f.marked[id] = location
	return nil
}

func TestArchivePendingUploadsAndMarks(t *testing.T) {
	service := newTestService(t)
	recordingPath := filepath.Join(t.TempDir(), "segment.mp4")
	if err := writeTestRecording(recordingPath); err != nil {
		t.Fatal(err)
	}
	index := &fakeArchiveIndex{
		segments: []ArchiveSegment{{ID: "recording-1", Path: "cam/segment.mp4"}},
		paths:    map[string]string{"cam/segment.mp4": recordingPath},
		marked:   make(map[string]string),
	}
	var uploadedID, uploadedKey, uploadedPath string
	service.archiveFile = func(_ context.Context, id, key, path string) (string, error) {
		uploadedID, uploadedKey, uploadedPath = id, key, path
		return "gdrive:file-1", nil
	}

	stats, err := service.ArchivePending(context.Background(), index, 10)
	if err != nil {
		t.Fatalf("ArchivePending: %v", err)
	}
	if stats.Uploaded != 1 || stats.Failed != 0 || stats.Skipped != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if uploadedID != "recording-1" || uploadedKey != "cam/segment.mp4" || uploadedPath != recordingPath {
		t.Fatalf("unexpected upload: id=%q key=%q path=%q", uploadedID, uploadedKey, uploadedPath)
	}
	if index.marked["recording-1"] != "gdrive:file-1" {
		t.Fatalf("unexpected archive location: %+v", index.marked)
	}
}

func TestArchivePendingMarksMissingFileSkipped(t *testing.T) {
	service := newTestService(t)
	index := &fakeArchiveIndex{
		segments: []ArchiveSegment{{ID: "recording-1", Path: "missing.mp4"}},
		paths:    map[string]string{"missing.mp4": filepath.Join(t.TempDir(), "missing.mp4")},
		marked:   make(map[string]string),
	}
	service.archiveFile = func(context.Context, string, string, string) (string, error) {
		t.Fatal("missing file must not be uploaded")
		return "", nil
	}

	stats, err := service.ArchivePending(context.Background(), index, 10)
	if err != nil {
		t.Fatalf("ArchivePending: %v", err)
	}
	if stats.Skipped != 1 || stats.Uploaded != 0 || stats.Failed != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if index.marked["recording-1"] != LocationMissing {
		t.Fatalf("unexpected skip location: %+v", index.marked)
	}
}

func TestArchivePendingIsSingleFlight(t *testing.T) {
	service := newTestService(t)
	service.archiveMu.Lock()
	defer service.archiveMu.Unlock()

	_, err := service.ArchivePending(context.Background(), &fakeArchiveIndex{}, 1)
	if !errors.Is(err, ErrArchiveInProgress) {
		t.Fatalf("got %v, want ErrArchiveInProgress", err)
	}
}

func writeDriveJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.WriteString(w, body); err != nil {
		t.Errorf("write response: %v", err)
	}
}
