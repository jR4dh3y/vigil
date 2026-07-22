package recording

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/nvr/nvr/server/internal/store"
)

func setupTestService(t *testing.T) (*Service, *store.Queries) {
	t.Helper()
	dir := t.TempDir()
	db, err := store.Open(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	q := store.New(db)

	// Parent camera required by FK.
	_, err = q.CreateCamera(context.Background(), store.CreateCameraParams{
		ID:          "550e8400-e29b-41d4-a716-446655440000",
		Name:        "Test Cam",
		Driver:      "generic-rtsp",
		Host:        "10.0.0.1",
		Username:    "",
		PasswordEnc: "",
		Enabled:     1,
		Status:      "unknown",
	})
	if err != nil {
		t.Fatalf("create camera: %v", err)
	}

	recDir := filepath.Join(dir, "recordings")
	svc := NewService(q, Config{RecordingsDir: recDir, RetentionDays: 7})
	return svc, q
}

func TestIndexSegmentAndListRange(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()
	camID := "550e8400-e29b-41d4-a716-446655440000"

	start := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	seg, err := svc.IndexSegment(ctx, camID,
		filepath.Join(svc.RecordingsDir(), camID, "2026-03-15", "12-00-00.mp4"),
		start, 60, 1024, "h264")
	if err != nil {
		t.Fatalf("IndexSegment: %v", err)
	}
	if seg.ID == "" || seg.CameraID != camID {
		t.Fatalf("unexpected segment: %+v", seg)
	}
	// Path should be relative to RecordingsDir.
	if seg.Path != camID+"/2026-03-15/12-00-00.mp4" {
		t.Fatalf("path: got %q", seg.Path)
	}
	if seg.Codec == nil || *seg.Codec != "h264" {
		t.Fatalf("codec: %v", seg.Codec)
	}

	// Outside range → empty.
	empty, err := svc.List(ctx, camID,
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if len(empty.Segments) != 0 {
		t.Fatalf("expected no segments, got %d", len(empty.Segments))
	}

	// Inclusive range containing started_at.
	got, err := svc.List(ctx, camID,
		time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 15, 23, 59, 59, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got.Segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(got.Segments))
	}
	if len(got.Coverage) != 1 {
		t.Fatalf("expected 1 coverage bar, got %d", len(got.Coverage))
	}
	if !got.Coverage[0].Start.Equal(start) {
		t.Fatalf("coverage start: %v", got.Coverage[0].Start)
	}
	wantEnd := start.Add(60 * time.Second)
	if !got.Coverage[0].End.Equal(wantEnd) {
		t.Fatalf("coverage end: got %v want %v", got.Coverage[0].End, wantEnd)
	}
}

func TestCameraIDFromPathName(t *testing.T) {
	id, ok := CameraIDFromPathName("cam_550e8400e29b41d4a716446655440000")
	if !ok || id != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("got %q ok=%v", id, ok)
	}
	id, ok = CameraIDFromPathName("550e8400-e29b-41d4-a716-446655440000")
	if !ok || id != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("uuid direct: %q ok=%v", id, ok)
	}
	if _, ok := CameraIDFromPathName("not-a-path"); ok {
		t.Fatal("expected failure")
	}
}

func TestPruneDeletesOldRows(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()
	camID := "550e8400-e29b-41d4-a716-446655440000"

	old := time.Now().UTC().AddDate(0, 0, -30)
	if _, err := svc.IndexSegment(ctx, camID, "old/seg.mp4", old, 60, 1, ""); err != nil {
		t.Fatalf("index old: %v", err)
	}
	recent := time.Now().UTC().Add(-time.Hour)
	if _, err := svc.IndexSegment(ctx, camID, "new/seg.mp4", recent, 60, 1, ""); err != nil {
		t.Fatalf("index new: %v", err)
	}

	n, err := svc.Prune(ctx)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 pruned, got %d", n)
	}

	list, err := svc.List(ctx, camID, time.Now().UTC().AddDate(0, 0, -60), time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Segments) != 1 || list.Segments[0].Path != "new/seg.mp4" {
		t.Fatalf("remaining: %+v", list.Segments)
	}
}
