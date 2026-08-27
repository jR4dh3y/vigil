package recording

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testCameraID = "550e8400-e29b-41d4-a716-446655440000"

// indexSegmentWithFile indexes a segment and creates its local MP4 with the
// supplied content and modification time, mirroring MediaMTX output.
func indexSegmentWithFile(
	t *testing.T,
	svc *Service,
	startedAt time.Time,
	durationSec float64,
	content string,
	modTime time.Time,
) Segment {
	t.Helper()
	name := startedAt.UTC().Format("15-04-05") + ".mp4"
	day := startedAt.UTC().Format("2006-01-02")
	abs := filepath.Join(svc.RecordingsDir(), testCameraID, day, name)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.Chtimes(abs, modTime, modTime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
	seg, err := svc.IndexSegment(
		context.Background(),
		testCameraID,
		abs,
		startedAt,
		durationSec,
		int64(len(content)),
		"h264",
	)
	if err != nil {
		t.Fatalf("IndexSegment: %v", err)
	}
	return seg
}

func requireFileState(t *testing.T, svc *Service, seg Segment, wantExists bool) {
	t.Helper()
	path, exists, err := svc.LocalPath(seg)
	if err != nil {
		t.Fatalf("LocalPath: %v", err)
	}
	if exists != wantExists {
		t.Fatalf("file %s exists = %v, want %v", path, exists, wantExists)
	}
}

func TestEnforceLocalLimitsNoopWhenDisabled(t *testing.T) {
	svc, _ := setupTestService(t)
	now := time.Now().UTC()
	indexSegmentWithFile(t, svc, now.Add(-2*time.Hour), 60, "old", now.Add(-time.Hour))

	stats, err := svc.EnforceLocalLimits(context.Background(), LocalLimits{})
	if err != nil {
		t.Fatalf("EnforceLocalLimits: %v", err)
	}
	if stats.Removed != 0 || stats.Examined != 0 {
		t.Fatalf("expected no-op stats, got %+v", stats)
	}
	segs, err := svc.ListUnarchived(context.Background(), 10)
	if err != nil || len(segs) != 1 {
		t.Fatalf("unarchived rows changed: n=%d err=%v", len(segs), err)
	}
}

func TestEnforceLocalLimitsDwellEvictsOldestAndMarksExpired(t *testing.T) {
	svc, q := setupTestService(t)
	now := time.Now().UTC()

	old := indexSegmentWithFile(t, svc, now.Add(-2*time.Hour), 60, "old", now.Add(-time.Hour))
	recent := indexSegmentWithFile(t, svc, now.Add(-2*time.Minute), 60, "recent", now.Add(-time.Minute))

	stats, err := svc.EnforceLocalLimits(context.Background(), LocalLimits{MaxDwell: time.Hour})
	if err != nil {
		t.Fatalf("EnforceLocalLimits: %v", err)
	}
	if stats.Removed != 1 || stats.FreedBytes != int64(len("old")) {
		t.Fatalf("stats = %+v, want one removal of %d bytes", stats, len("old"))
	}

	requireFileState(t, svc, old, false)
	requireFileState(t, svc, recent, true)

	row, err := q.GetRecording(context.Background(), old.ID)
	if err != nil {
		t.Fatalf("get evicted row: %v", err)
	}
	if !row.ArchiveLocation.Valid || row.ArchiveLocation.String != LocationExpired {
		t.Fatalf("archive location = %+v, want %q", row.ArchiveLocation, LocationExpired)
	}
	if !row.ArchivedAt.Valid {
		t.Fatalf("evicted row should be marked processed")
	}

	remaining, err := svc.ListUnarchived(context.Background(), 10)
	if err != nil {
		t.Fatalf("list unarchived: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != recent.ID {
		t.Fatalf("unarchived queue should only hold recent segment, got %+v", remaining)
	}
}

func TestEnforceLocalLimitsOverflowStopsBelowThreshold(t *testing.T) {
	svc, _ := setupTestService(t)
	now := time.Now().UTC()

	first := indexSegmentWithFile(t, svc, now.Add(-3*time.Minute), 60, "first", now.Add(-2*time.Minute))
	second := indexSegmentWithFile(t, svc, now.Add(-2*time.Minute), 60, "second", now.Add(-90*time.Second))

	calls := 0
	svc.usageFn = func(ctx context.Context) (float64, error) {
		calls++
		// Above threshold on first inspection; below after one eviction.
		if calls == 1 {
			return 95, nil
		}
		return 50, nil
	}

	stats, err := svc.EnforceLocalLimits(context.Background(), LocalLimits{MaxUsedPercent: 90})
	if err != nil {
		t.Fatalf("EnforceLocalLimits: %v", err)
	}
	if stats.Removed != 1 {
		t.Fatalf("removed = %d, want exactly the oldest segment (%+v)", stats.Removed, stats)
	}
	requireFileState(t, svc, first, false)
	requireFileState(t, svc, second, true)
}

func TestEnforceLocalLimitsOverflowEvictsUntilClear(t *testing.T) {
	svc, _ := setupTestService(t)
	now := time.Now().UTC()

	a := indexSegmentWithFile(t, svc, now.Add(-4*time.Minute), 60, "a", now.Add(-3*time.Minute))
	b := indexSegmentWithFile(t, svc, now.Add(-3*time.Minute), 60, "b", now.Add(-2*time.Minute))

	svc.usageFn = func(ctx context.Context) (float64, error) { return 99, nil }

	stats, err := svc.EnforceLocalLimits(context.Background(), LocalLimits{MaxUsedPercent: 90})
	if err != nil {
		t.Fatalf("EnforceLocalLimits: %v", err)
	}
	if stats.Removed != 2 {
		t.Fatalf("removed = %d, want both candidates (%+v)", stats.Removed, stats)
	}
	requireFileState(t, svc, a, false)
	requireFileState(t, svc, b, false)
}

func TestEnforceLocalLimitsRespectsSettleAge(t *testing.T) {
	svc, q := setupTestService(t)
	now := time.Now().UTC()

	// Ended long ago (dwell-due by start/duration) but written moments ago.
	freshWrite := indexSegmentWithFile(t, svc, now.Add(-2*time.Hour), 60, "fresh-write", now)

	stats, err := svc.EnforceLocalLimits(context.Background(), LocalLimits{MaxDwell: time.Hour})
	if err != nil {
		t.Fatalf("EnforceLocalLimits: %v", err)
	}
	if stats.BusySkipped != 1 || stats.Removed != 0 {
		t.Fatalf("stats = %+v, want busy skip without removal", stats)
	}
	requireFileState(t, svc, freshWrite, true)
	row, err := q.GetRecording(context.Background(), freshWrite.ID)
	if err != nil {
		t.Fatalf("get row: %v", err)
	}
	if row.ArchiveLocation.Valid {
		t.Fatalf("busy row must stay unarchived, got %q", row.ArchiveLocation.String)
	}
}

func TestEnforceLocalLimitsUsageErrorIsFatal(t *testing.T) {
	svc, _ := setupTestService(t)
	now := time.Now().UTC()
	indexSegmentWithFile(t, svc, now.Add(-time.Minute), 60, "seg", now.Add(-30*time.Second))

	boom := errors.New("statfs unavailable")
	svc.usageFn = func(ctx context.Context) (float64, error) { return 0, boom }

	if _, err := svc.EnforceLocalLimits(context.Background(), LocalLimits{MaxUsedPercent: 50}); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestSegmentCompleteNotifiesListener(t *testing.T) {
	svc, _ := setupTestService(t)

	notified := make(chan Segment, 1)
	svc.SetSegmentListener(func(seg Segment) {
		select {
		case notified <- seg:
		default:
		}
	})

	handler := svc.SegmentCompleteHandler()
	body := `{"path":"` + testCameraID + `","file_path":"` +
		filepath.Join(svc.RecordingsDir(), testCameraID, "2026-08-01", "10-00-00.mp4") +
		`","duration_sec":60,"size_bytes":2048}`
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/internal/mediamtx/segment-complete", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("hook status = %d, body %s", rec.Code, rec.Body.String())
	}

	select {
	case seg := <-notified:
		if seg.CameraID != testCameraID || seg.SizeBytes != 2048 {
			t.Fatalf("unexpected notified segment: %+v", seg)
		}
	case <-time.After(time.Second):
		t.Fatalf("listener was not called for completed segment")
	}
}

func TestSegmentCompleteWithoutListenerStillIndexes(t *testing.T) {
	svc, q := setupTestService(t)

	handler := svc.SegmentCompleteHandler()
	body := `{"path":"` + testCameraID + `","file_path":"` +
		filepath.Join(svc.RecordingsDir(), testCameraID, "2026-08-01", "11-00-00.mp4") +
		`","duration_sec":60}`
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/internal/mediamtx/segment-complete", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("hook status = %d, body %s", rec.Code, rec.Body.String())
	}
	rows, err := q.ListUnarchivedRecordings(context.Background(), 10)
	if err != nil || len(rows) != 1 {
		t.Fatalf("indexed rows = %d, err %v, want 1", len(rows), err)
	}
}

func TestEnforceLocalLimitsDwellSkipsNonExpiredAndContinues(t *testing.T) {
	svc, q := setupTestService(t)
	now := time.Now().UTC()

	// Regression test: a long-dwell early segment (not yet expired) followed by
	// a short-dwell segment that IS expired. The dwell scan must continue past
	// the non-expired segment and still evict the short expired one.
	longNotExpired := indexSegmentWithFile(t, svc, now.Add(-30*time.Minute), 60, "long", now.Add(-29*time.Minute))
	shortExpired := indexSegmentWithFile(t, svc, now.Add(-3*time.Hour), 60, "short", now.Add(-2*time.Hour))

	stats, err := svc.EnforceLocalLimits(context.Background(), LocalLimits{MaxDwell: time.Hour})
	if err != nil {
		t.Fatalf("EnforceLocalLimits: %v", err)
	}
	if stats.Removed != 1 || stats.FreedBytes != int64(len("short")) {
		t.Fatalf("stats = %+v, want one removal of short segment (%d bytes)", stats, len("short"))
	}

	requireFileState(t, svc, longNotExpired, true)
	requireFileState(t, svc, shortExpired, false)

	row, err := q.GetRecording(context.Background(), shortExpired.ID)
	if err != nil {
		t.Fatalf("get evicted row: %v", err)
	}
	if !row.ArchiveLocation.Valid || row.ArchiveLocation.String != LocationExpired {
		t.Fatalf("archive location = %+v, want %q", row.ArchiveLocation, LocationExpired)
	}

	remaining, err := svc.ListUnarchived(context.Background(), 10)
	if err != nil {
		t.Fatalf("list unarchived: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != longNotExpired.ID {
		t.Fatalf("unarchived queue should only hold non-expired segment, got %+v", remaining)
	}
}
