package recording

import (
	"context"
	"errors"
	"os"
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

func TestIndexSegmentIsIdempotentByPath(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()
	const camID = "550e8400-e29b-41d4-a716-446655440000"
	path := filepath.Join(svc.RecordingsDir(), "cam_550e8400e29b41d4a716446655440000", "2026-08-12", "14-00-00.mp4")
	started := time.Date(2026, 8, 12, 14, 0, 0, 0, time.UTC)

	first, err := svc.IndexSegment(ctx, camID, path, started, 60, 100, "")
	if err != nil {
		t.Fatalf("first IndexSegment: %v", err)
	}
	second, err := svc.IndexSegment(ctx, camID, path, started, 61, 200, "h264")
	if err != nil {
		t.Fatalf("second IndexSegment: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("duplicate path changed id: %q != %q", second.ID, first.ID)
	}
	if second.DurationSec != 61 || second.SizeBytes != 200 {
		t.Fatalf("segment was not refreshed: %+v", second)
	}

	got, err := svc.List(ctx, camID, started.Add(-time.Minute), started.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Segments) != 1 {
		t.Fatalf("got %d segments, want 1", len(got.Segments))
	}
}

func TestFindAtAndLocalPath(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()
	const camID = "550e8400-e29b-41d4-a716-446655440000"
	start := time.Date(2026, 8, 12, 14, 0, 0, 0, time.UTC)
	local := filepath.Join(svc.RecordingsDir(), "cam", "segment.mp4")
	if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(local, []byte("mp4"), 0o600); err != nil {
		t.Fatal(err)
	}
	segment, err := svc.IndexSegment(ctx, camID, local, start, 60, 3, "h264")
	if err != nil {
		t.Fatal(err)
	}

	found, err := svc.FindAt(ctx, camID, start.Add(25*time.Second))
	if err != nil || found.ID != segment.ID {
		t.Fatalf("FindAt = %+v, %v", found, err)
	}
	path, available, err := svc.LocalPath(found)
	if err != nil || !available || path != local {
		t.Fatalf("LocalPath = %q, %v, %v", path, available, err)
	}
	if _, err := svc.FindAt(ctx, camID, start.Add(2*time.Minute)); !errors.Is(err, ErrOutsideRecording) {
		t.Fatalf("gap lookup error = %v, want ErrOutsideRecording", err)
	}
	if err := os.Remove(local); err != nil {
		t.Fatal(err)
	}
	_, available, err = svc.LocalPath(found)
	if err != nil || available {
		t.Fatalf("missing LocalPath available=%v err=%v", available, err)
	}
}

func TestReconcileDiskBackfillsAndSkipsActiveFiles(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()
	const camID = "550e8400-e29b-41d4-a716-446655440000"
	cameraDir := filepath.Join(svc.RecordingsDir(), "cam_550e8400e29b41d4a716446655440000", "2026-08-12")
	if err := os.MkdirAll(cameraDir, 0o755); err != nil {
		t.Fatal(err)
	}
	completed := filepath.Join(cameraDir, "14-00-00-000000.mp4")
	active := filepath.Join(cameraDir, "14-01-00-000000.mp4")
	if err := os.WriteFile(completed, []byte("completed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(completed, time.Now().Add(-time.Minute), time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(active, []byte("active"), 0o600); err != nil {
		t.Fatal(err)
	}

	stats, err := svc.ReconcileDisk(ctx, 5*time.Second)
	if err != nil {
		t.Fatalf("ReconcileDisk: %v", err)
	}
	if stats.Indexed != 1 || stats.Skipped != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	segments, err := svc.List(
		ctx,
		camID,
		time.Date(2026, 8, 12, 13, 59, 0, 0, time.UTC),
		time.Date(2026, 8, 12, 14, 2, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(segments.Segments) != 1 {
		t.Fatalf("got %d segments, want 1", len(segments.Segments))
	}

	stats, err = svc.ReconcileDisk(ctx, 5*time.Second)
	if err != nil {
		t.Fatalf("second ReconcileDisk: %v", err)
	}
	if stats.Existing != 1 || stats.Indexed != 0 {
		t.Fatalf("reconciliation was not idempotent: %+v", stats)
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

func TestStartedAtFromMediaMTXFileName(t *testing.T) {
	got, ok := startedAtFromFileName(
		"/recordings/cam_550e8400e29b41d4a716446655440000/2026-08-12/14-53-34-741969.mp4",
	)
	if !ok {
		t.Fatal("expected timestamp to parse")
	}
	want := time.Date(2026, 8, 12, 14, 53, 34, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAbsolutePathRejectsTraversal(t *testing.T) {
	svc, _ := setupTestService(t)
	if _, err := svc.AbsolutePath("../etc/passwd"); err == nil {
		t.Fatal("expected traversal rejection")
	}
	if _, err := svc.AbsolutePath("/abs/path"); err == nil {
		t.Fatal("expected absolute rejection")
	}
	got, err := svc.AbsolutePath("cam/seg.mp4")
	if err != nil {
		t.Fatalf("AbsolutePath: %v", err)
	}
	want := filepath.Join(svc.RecordingsDir(), "cam", "seg.mp4")
	absWant, err := filepath.Abs(want)
	if err != nil {
		t.Fatal(err)
	}
	if got != absWant {
		t.Fatalf("got %q want %q", got, absWant)
	}
}

func TestAbsolutePathRejectsSymlinkEscape(t *testing.T) {
	svc, _ := setupTestService(t)
	if err := os.MkdirAll(svc.RecordingsDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.mp4")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(svc.RecordingsDir(), "segment.mp4")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AbsolutePath("segment.mp4"); err == nil {
		t.Fatal("expected symlink escape rejection")
	}
}

func TestListUnarchivedAndMarkArchived(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()
	camID := "550e8400-e29b-41d4-a716-446655440000"

	start := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	seg, err := svc.IndexSegment(ctx, camID, "cam/seg.mp4", start, 60, 1024, "h264")
	if err != nil {
		t.Fatalf("IndexSegment: %v", err)
	}

	list, err := svc.ListUnarchived(ctx, 10)
	if err != nil {
		t.Fatalf("ListUnarchived: %v", err)
	}
	if len(list) != 1 || list[0].ID != seg.ID {
		t.Fatalf("unarchived: %+v", list)
	}

	if err := svc.MarkArchived(ctx, seg.ID, "gdrive:file123"); err != nil {
		t.Fatalf("MarkArchived: %v", err)
	}
	list, err = svc.ListUnarchived(ctx, 10)
	if err != nil {
		t.Fatalf("ListUnarchived after mark: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty unarchived, got %+v", list)
	}
	if err := svc.MarkArchived(ctx, "missing-recording", "gdrive:file123"); err == nil {
		t.Fatal("expected missing recording error")
	}
}

func TestPruneArchivedPreservesPendingRows(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()
	camID := "550e8400-e29b-41d4-a716-446655440000"
	old := time.Now().UTC().AddDate(0, 0, -30)

	archived, err := svc.IndexSegment(ctx, camID, "old/archived.mp4", old, 60, 1, "")
	if err != nil {
		t.Fatalf("index archived: %v", err)
	}
	if _, err := svc.IndexSegment(ctx, camID, "old/pending.mp4", old.Add(time.Minute), 60, 1, ""); err != nil {
		t.Fatalf("index pending: %v", err)
	}
	if err := svc.MarkArchived(ctx, archived.ID, "skipped:missing"); err != nil {
		t.Fatalf("mark archived: %v", err)
	}

	n, err := svc.PruneArchived(ctx)
	if err != nil {
		t.Fatalf("PruneArchived: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 archived row pruned, got %d", n)
	}
	pending, err := svc.ListUnarchived(ctx, 10)
	if err != nil {
		t.Fatalf("ListUnarchived: %v", err)
	}
	if len(pending) != 1 || pending[0].Path != "old/pending.mp4" {
		t.Fatalf("unexpected pending rows: %+v", pending)
	}
}

func TestPrunePreservesDriveArchiveMetadata(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()
	const camID = "550e8400-e29b-41d4-a716-446655440000"
	old := time.Now().UTC().AddDate(0, 0, -30)
	archived, err := svc.IndexSegment(ctx, camID, "old/drive.mp4", old, 60, 1, "h264")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.MarkArchived(ctx, archived.ID, "gdrive:file123"); err != nil {
		t.Fatal(err)
	}
	if n, err := svc.Prune(ctx); err != nil || n != 0 {
		t.Fatalf("Prune = %d, %v; Drive metadata must be preserved", n, err)
	}
	found, err := svc.FindAt(ctx, camID, old.Add(30*time.Second))
	if err != nil || found.ID != archived.ID {
		t.Fatalf("archived recording missing after prune: %+v, %v", found, err)
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
