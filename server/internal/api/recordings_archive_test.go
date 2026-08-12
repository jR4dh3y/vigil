package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/camera"
	"github.com/nvr/nvr/server/internal/media"
	"github.com/nvr/nvr/server/internal/recording"
	"github.com/nvr/nvr/server/internal/store"
)

const archiveTestCameraID = "550e8400-e29b-41d4-a716-446655440000"

type archiveCameraReader struct{}

func (archiveCameraReader) Get(context.Context, string) (camera.Camera, error) {
	return camera.Camera{ID: archiveTestCameraID, Enabled: true}, nil
}

func (archiveCameraReader) LiveSourceURL(context.Context, string) (string, error) {
	return "", nil
}

type fakeDrivePlayback struct {
	fileID    string
	byteRange string
	calls     int
}

func (f *fakeDrivePlayback) Download(_ context.Context, fileID, byteRange string) (*http.Response, error) {
	f.fileID = fileID
	f.byteRange = byteRange
	f.calls++
	return &http.Response{
		StatusCode: http.StatusPartialContent,
		Header: http.Header{
			"Content-Type":   []string{"video/mp4"},
			"Content-Length": []string{"11"},
			"Content-Range":  []string{"bytes 0-10/100"},
		},
		Body: io.NopCloser(bytes.NewBufferString("hello drive")),
	}, nil
}

func setupArchivePlaybackServer(t *testing.T) (*Server, recording.Segment) {
	t.Helper()
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatal(err)
	}
	queries := store.New(db)
	if _, err := queries.CreateCamera(context.Background(), store.CreateCameraParams{
		ID: archiveTestCameraID, Name: "Archive Cam", Driver: "generic-rtsp", Host: "10.0.0.1",
		Username: "", PasswordEnc: "", Enabled: 1, Status: "online",
	}); err != nil {
		t.Fatal(err)
	}
	recordingSvc := recording.NewService(queries, recording.Config{
		RecordingsDir: filepath.Join(t.TempDir(), "recordings"),
		RetentionDays: 7,
	})
	segment, err := recordingSvc.IndexSegment(
		context.Background(), archiveTestCameraID, "cam/segment.mp4",
		time.Date(2026, 8, 12, 14, 0, 0, 0, time.UTC), 60, 100, "h264",
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := recordingSvc.MarkArchived(context.Background(), segment.ID, "gdrive:drive-file-1"); err != nil {
		t.Fatal(err)
	}
	mediaSvc := media.NewService(media.Config{PlaybackURL: "https://playback.example.test"}, archiveCameraReader{})
	return &Server{Recording: recordingSvc, Media: mediaSvc}, segment
}

func TestPlaybackFallsBackToDriveWhenLocalFileIsGone(t *testing.T) {
	server, segment := setupArchivePlaybackServer(t)
	server.DrivePlayback = &fakeDrivePlayback{}
	body := bytes.NewBufferString(`{"start":"2026-08-12T14:00:10Z","durationSec":60}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cameras/"+archiveTestCameraID+"/playback", body)
	req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: "user", Role: auth.RoleViewer}))
	rr := httptest.NewRecorder()
	server.PostCameraPlayback(rr, req, uuid.MustParse(archiveTestCameraID))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var session PlaybackSession
	if err := json.Unmarshal(rr.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.Source != Gdrive || session.RecordingId != segment.ID || session.StartOffsetSec != 10 {
		t.Fatalf("unexpected playback session: %+v", session)
	}
	if session.PlaybackUrl == "" || session.Token == "" {
		t.Fatalf("missing Drive playback credentials: %+v", session)
	}
}

func TestRecordingContentStreamsDriveRange(t *testing.T) {
	server, segment := setupArchivePlaybackServer(t)
	drive := &fakeDrivePlayback{}
	server.DrivePlayback = drive
	session, err := server.Media.IssueArchivedPlayback(context.Background(), archiveTestCameraID, segment.ID)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, session.PlaybackURL, nil)
	req.Header.Set("Range", "bytes=0-10")
	rr := httptest.NewRecorder()
	server.GetRecordingContent(rr, req, uuid.MustParse(segment.ID), GetRecordingContentParams{Token: session.Token})
	if rr.Code != http.StatusPartialContent || rr.Body.String() != "hello drive" {
		t.Fatalf("status=%d body=%q", rr.Code, rr.Body.String())
	}
	if drive.fileID != "drive-file-1" || drive.byteRange != "bytes=0-10" {
		t.Fatalf("Drive request file=%q range=%q", drive.fileID, drive.byteRange)
	}
	if rr.Header().Get("Content-Type") != "video/mp4" || rr.Header().Get("Content-Range") != "bytes 0-10/100" {
		t.Fatalf("unexpected response headers: %v", rr.Header())
	}

	bad := httptest.NewRecorder()
	server.GetRecordingContent(bad, httptest.NewRequest(http.MethodGet, "/", nil), uuid.MustParse(segment.ID), GetRecordingContentParams{Token: "bad"})
	if bad.Code != http.StatusUnauthorized || drive.calls != 1 {
		t.Fatalf("bad token status=%d Drive calls=%d", bad.Code, drive.calls)
	}
}
