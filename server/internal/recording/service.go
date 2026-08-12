package recording

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nvr/nvr/server/internal/store"
)

// DefaultRetentionDays is used when config does not specify a retention window.
const DefaultRetentionDays = 7

var (
	ErrNotFound         = errors.New("recording not found")
	ErrOutsideRecording = errors.New("requested time is outside the recording segment")
)

// Segment is a recorded media segment in the index.
type Segment struct {
	ID              string
	CameraID        string
	StartedAt       time.Time
	DurationSec     float64
	SizeBytes       int64
	Path            string
	Codec           *string
	ThumbnailPath   *string
	ArchivedAt      *time.Time
	ArchiveLocation *string
}

// CoverageBar is a continuous coverage window for the timeline UI.
type CoverageBar struct {
	Start time.Time
	End   time.Time
}

// ListResult is the timeline query response: segments plus coverage bars.
type ListResult struct {
	Segments []Segment
	Coverage []CoverageBar
}

// Config holds recording service options.
type Config struct {
	// RecordingsDir is the absolute or relative root for recording files.
	RecordingsDir string
	// RetentionDays is how long to keep segments (Prune). 0 uses DefaultRetentionDays.
	RetentionDays int
}

// Service indexes recording segments and answers timeline queries.
type Service struct {
	q             *store.Queries
	recordingsDir string
	retentionDays int
	reconcileMu   sync.Mutex
}

// SetRecordingsDir updates the root used to relativize segment paths.
func (s *Service) SetRecordingsDir(dir string) {
	s.recordingsDir = strings.TrimSpace(dir)
}

// NewService constructs a recording Service.
func NewService(q *store.Queries, cfg Config) *Service {
	days := cfg.RetentionDays
	if days <= 0 {
		days = DefaultRetentionDays
	}
	return &Service{
		q:             q,
		recordingsDir: strings.TrimSpace(cfg.RecordingsDir),
		retentionDays: days,
	}
}

// Get returns a recording segment by ID.
func (s *Service) Get(ctx context.Context, id string) (Segment, error) {
	row, err := s.q.GetRecording(ctx, strings.TrimSpace(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Segment{}, ErrNotFound
		}
		return Segment{}, fmt.Errorf("get recording: %w", err)
	}
	return toSegment(row)
}

// FindAt returns the segment covering at for a camera. The lookup starts with
// the latest segment at or before at, then verifies that at is inside it.
func (s *Service) FindAt(ctx context.Context, cameraID string, at time.Time) (Segment, error) {
	row, err := s.q.GetRecordingAtOrBefore(ctx, store.GetRecordingAtOrBeforeParams{
		CameraID: strings.TrimSpace(cameraID),
		AtTs:     formatTime(at.UTC()),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Segment{}, ErrNotFound
		}
		return Segment{}, fmt.Errorf("find recording at time: %w", err)
	}
	segment, err := toSegment(row)
	if err != nil {
		return Segment{}, err
	}
	end := segment.StartedAt.Add(time.Duration(segment.DurationSec * float64(time.Second)))
	if at.Before(segment.StartedAt) || at.After(end) {
		return Segment{}, ErrOutsideRecording
	}
	return segment, nil
}

// LocalPath returns the safe absolute path and whether the recording exists as
// a regular local file.
func (s *Service) LocalPath(segment Segment) (string, bool, error) {
	abs, err := s.AbsolutePath(segment.Path)
	if err != nil {
		return "", false, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return abs, false, nil
		}
		return "", false, fmt.Errorf("stat recording: %w", err)
	}
	return abs, info.Mode().IsRegular(), nil
}

// List returns segments and coverage bars for cameraID in [from, to] (inclusive).
// Coverage is 1:1 with segments for phase 4 (adjacent merge is optional later).
func (s *Service) List(ctx context.Context, cameraID string, from, to time.Time) (ListResult, error) {
	if from.After(to) {
		from, to = to, from
	}
	rows, err := s.q.ListRecordingsByCameraRange(ctx, store.ListRecordingsByCameraRangeParams{
		CameraID: cameraID,
		FromTs:   formatTime(from),
		ToTs:     formatTime(to),
	})
	if err != nil {
		return ListResult{}, fmt.Errorf("list recordings: %w", err)
	}

	segments := make([]Segment, 0, len(rows))
	coverage := make([]CoverageBar, 0, len(rows))
	for _, row := range rows {
		seg, err := toSegment(row)
		if err != nil {
			return ListResult{}, err
		}
		segments = append(segments, seg)
		end := seg.StartedAt.Add(time.Duration(seg.DurationSec * float64(time.Second)))
		coverage = append(coverage, CoverageBar{Start: seg.StartedAt, End: end})
	}
	return ListResult{Segments: segments, Coverage: coverage}, nil
}

// IndexSegment inserts a recording row. path should be relative to RecordingsDir when possible.
func (s *Service) IndexSegment(ctx context.Context, cameraID, path string, startedAt time.Time, durationSec float64, sizeBytes int64, codec string) (Segment, error) {
	cameraID = strings.TrimSpace(cameraID)
	if cameraID == "" {
		return Segment{}, fmt.Errorf("camera_id is required")
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return Segment{}, fmt.Errorf("path is required")
	}
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	if durationSec < 0 {
		durationSec = 0
	}
	if sizeBytes < 0 {
		sizeBytes = 0
	}

	rel := s.relativizePath(path)
	var codecNS sql.NullString
	if c := strings.TrimSpace(codec); c != "" {
		codecNS = sql.NullString{String: c, Valid: true}
	}

	row, err := s.q.InsertRecording(ctx, store.InsertRecordingParams{
		ID:          uuid.NewString(),
		CameraID:    cameraID,
		StartedAt:   formatTime(startedAt.UTC()),
		DurationSec: durationSec,
		SizeBytes:   sizeBytes,
		Path:        rel,
		Codec:       codecNS,
	})
	if err != nil {
		return Segment{}, fmt.Errorf("insert recording: %w", err)
	}
	return toSegment(row)
}

// DeleteOlderThan removes non-Drive index rows before the cutoff. Successful
// Drive archive metadata is retained so old recordings remain on the timeline.
// File deletion is managed separately by MediaMTX.
func (s *Service) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	n, err := s.q.DeleteRecordingsOlderThan(ctx, formatTime(before.UTC()))
	if err != nil {
		return 0, fmt.Errorf("delete recordings older than: %w", err)
	}
	return n, nil
}

// Prune deletes non-Drive index rows older than the configured retention window.
func (s *Service) Prune(ctx context.Context) (int64, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -s.retentionDays)
	return s.DeleteOlderThan(ctx, cutoff)
}

// PruneArchived removes old rows that completed without a durable Drive object
// (for example skipped:missing). Successful Drive rows remain searchable.
func (s *Service) PruneArchived(ctx context.Context) (int64, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -s.retentionDays)
	n, err := s.q.DeleteArchivedRecordingsOlderThan(ctx, formatTime(cutoff))
	if err != nil {
		return 0, fmt.Errorf("delete archived recordings older than: %w", err)
	}
	return n, nil
}

// RetentionDays returns the configured retention window in days.
func (s *Service) RetentionDays() int {
	return s.retentionDays
}

// SetRetentionDays updates the retention window used by Prune.
// Values <= 0 fall back to DefaultRetentionDays.
func (s *Service) SetRetentionDays(days int) {
	if days <= 0 {
		days = DefaultRetentionDays
	}
	s.retentionDays = days
}

// RecordingsDir returns the configured recordings root.
func (s *Service) RecordingsDir() string {
	return s.recordingsDir
}

// AbsolutePath joins rel under RecordingsDir and rejects path traversal.
func (s *Service) AbsolutePath(rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", fmt.Errorf("empty path")
	}
	if s.recordingsDir == "" {
		return "", fmt.Errorf("recordings dir not configured")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("absolute path not allowed")
	}
	clean := filepath.Clean(rel)
	// Reject ".." segments after Clean (e.g. "../x" → still escapes).
	for _, part := range strings.Split(filepath.ToSlash(clean), "/") {
		if part == ".." {
			return "", fmt.Errorf("path traversal not allowed")
		}
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path traversal not allowed")
	}

	absRoot, err := filepath.Abs(s.recordingsDir)
	if err != nil {
		return "", fmt.Errorf("recordings dir: %w", err)
	}
	absPath, err := filepath.Abs(filepath.Join(absRoot, clean))
	if err != nil {
		return "", err
	}
	// Ensure result stays under root.
	if absPath != absRoot && !strings.HasPrefix(absPath, absRoot+string(os.PathSeparator)) {
		return "", fmt.Errorf("path outside recordings dir")
	}

	// Resolve symlinks for existing archive candidates. Without this check, a
	// symlink stored below RecordingsDir could expose an arbitrary host file.
	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return absPath, nil
		}
		return "", fmt.Errorf("resolve recordings dir: %w", err)
	}
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return absPath, nil
		}
		return "", fmt.Errorf("resolve recording path: %w", err)
	}
	relToRoot, err := filepath.Rel(realRoot, realPath)
	if err != nil {
		return "", fmt.Errorf("compare recording path: %w", err)
	}
	if relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("recording path resolves outside recordings dir")
	}
	return absPath, nil
}

// ListUnarchived returns up to limit segments that have not been archived yet.
func (s *Service) ListUnarchived(ctx context.Context, limit int) ([]Segment, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.q.ListUnarchivedRecordings(ctx, int64(limit))
	if err != nil {
		return nil, fmt.Errorf("list unarchived recordings: %w", err)
	}
	segments := make([]Segment, 0, len(rows))
	for _, row := range rows {
		seg, err := toSegment(row)
		if err != nil {
			return nil, err
		}
		segments = append(segments, seg)
	}
	return segments, nil
}

// MarkArchived sets archived_at to now (UTC RFC3339) and archive_location.
func (s *Service) MarkArchived(ctx context.Context, id, location string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("id is required")
	}
	location = strings.TrimSpace(location)
	if location == "" {
		return fmt.Errorf("archive location is required")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	n, err := s.q.MarkRecordingArchived(ctx, store.MarkRecordingArchivedParams{
		ArchivedAt:      sql.NullString{String: now, Valid: true},
		ArchiveLocation: sql.NullString{String: location, Valid: true},
		ID:              id,
	})
	if err != nil {
		return fmt.Errorf("mark recording archived: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("mark recording archived: recording %q not found", id)
	}
	return nil
}

// relativizePath stores paths relative to RecordingsDir when possible.
func (s *Service) relativizePath(path string) string {
	path = filepath.Clean(path)
	if s.recordingsDir == "" {
		return filepath.ToSlash(path)
	}
	root := filepath.Clean(s.recordingsDir)
	// Try absolute comparison.
	absPath, err1 := filepath.Abs(path)
	absRoot, err2 := filepath.Abs(root)
	if err1 == nil && err2 == nil {
		if rel, err := filepath.Rel(absRoot, absPath); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
	}
	// Already relative under root prefix.
	if strings.HasPrefix(path, root+string(os.PathSeparator)) || path == root {
		if rel, err := filepath.Rel(root, path); err == nil {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(path)
}

func toSegment(row store.Recording) (Segment, error) {
	started, err := parseTime(row.StartedAt)
	if err != nil {
		return Segment{}, fmt.Errorf("parse started_at %q: %w", row.StartedAt, err)
	}
	seg := Segment{
		ID:          row.ID,
		CameraID:    row.CameraID,
		StartedAt:   started,
		DurationSec: row.DurationSec,
		SizeBytes:   row.SizeBytes,
		Path:        row.Path,
	}
	if row.Codec.Valid {
		c := row.Codec.String
		seg.Codec = &c
	}
	if row.ThumbnailPath.Valid {
		t := row.ThumbnailPath.String
		seg.ThumbnailPath = &t
	}
	if row.ArchivedAt.Valid && row.ArchivedAt.String != "" {
		if t, err := parseTime(row.ArchivedAt.String); err == nil {
			seg.ArchivedAt = &t
		}
	}
	if row.ArchiveLocation.Valid && row.ArchiveLocation.String != "" {
		loc := row.ArchiveLocation.String
		seg.ArchiveLocation = &loc
	}
	return seg, nil
}

// formatTime stores timestamps in RFC3339 (UTC) for stable range queries.
func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// parseTime parses RFC3339 and common SQLite / MediaMTX time formats.
func parseTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02_15-04-05",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.UTC); err == nil {
			return t.UTC(), nil
		}
	}
	// Unix seconds / milliseconds as string.
	if n, err := parseFlexibleUnix(s); err == nil {
		return n, nil
	}
	return time.Time{}, fmt.Errorf("parse time %q", s)
}

func parseFlexibleUnix(s string) (time.Time, error) {
	var n int64
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return time.Time{}, err
	}
	// Heuristic: ms vs s.
	if n > 1_000_000_000_000 {
		return time.UnixMilli(n).UTC(), nil
	}
	return time.Unix(n, 0).UTC(), nil
}
