package media

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/nvr/nvr/server/internal/camera"
)

type playbackCameraReader struct {
	camera camera.Camera
}

func (r playbackCameraReader) Get(context.Context, string) (camera.Camera, error) {
	return r.camera, nil
}

func (playbackCameraReader) LiveSourceURL(context.Context, string) (string, error) {
	return "", nil
}

func TestIssuePlaybackReturnsBrowserCompatibleMP4(t *testing.T) {
	const cameraID = "550e8400-e29b-41d4-a716-446655440000"
	svc := NewService(Config{
		PlaybackURL: "https://playback.example.test",
	}, playbackCameraReader{camera: camera.Camera{ID: cameraID}})

	session, err := svc.IssuePlayback(
		context.Background(),
		cameraID,
		time.Date(2026, 8, 12, 14, 0, 0, 0, time.UTC),
		60,
	)
	if err != nil {
		t.Fatalf("IssuePlayback: %v", err)
	}
	playbackURL, err := url.Parse(session.PlaybackURL)
	if err != nil {
		t.Fatalf("parse playback URL: %v", err)
	}
	query := playbackURL.Query()
	if got := query.Get("path"); got != PathName(cameraID) {
		t.Fatalf("path = %q, want %q", got, PathName(cameraID))
	}
	if got := query.Get("format"); got != "mp4" {
		t.Fatalf("format = %q, want mp4", got)
	}
	if query.Get("token") == "" {
		t.Fatal("playback token is missing")
	}
}
