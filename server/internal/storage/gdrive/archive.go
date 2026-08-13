package gdrive

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// ArchiveSegment is the minimal recording metadata needed for archive upload.
type ArchiveSegment struct {
	ID        string
	Path      string
	CameraID  string
	StartedAt time.Time
}

// ArchiveIndex provides recording metadata for archive uploads.
// Implemented by an adapter over recording.Service (avoids gdrive → recording import).
type ArchiveIndex interface {
	ListUnarchived(ctx context.Context, limit int) ([]ArchiveSegment, error)
	AbsolutePath(rel string) (string, error)
	MarkArchived(ctx context.Context, id, location string) error
	DeleteLocal(ctx context.Context, id, rel string) error
}

// ArchiveStats summarizes a batch archive run.
type ArchiveStats struct {
	Uploaded     int
	Deleted      int
	DeleteFailed int
	Failed       int
	Skipped      int
}

// DefaultArchiveBatchLimit is used when limit <= 0.
// This limit is aligned with the API's maximum validation to ensure consistent batch sizes.
const DefaultArchiveBatchLimit = 50

// ErrArchiveInProgress is returned when a cron or API archive pass already owns
// the service's single-flight lock.
var ErrArchiveInProgress = errors.New("google drive archive already in progress")

// Connected reports whether stored Drive credentials are configured and decryptable.
func (s *Service) Connected(ctx context.Context) bool {
	if s == nil || s.q == nil {
		return false
	}
	st, err := s.Status(ctx)
	return err == nil && st.Connected
}

// ArchiveLocalFile opens localPath, uploads it as objectKey, and returns location
// "gdrive:{fileID}". recordingID is stored as Drive app metadata for idempotency.
func (s *Service) ArchiveLocalFile(ctx context.Context, recordingID, objectKey, localPath string) (location string, err error) {
	f, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open %q: %w", localPath, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("stat %q: %w", localPath, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%q is a directory", localPath)
	}

	fileID, err := s.Upload(ctx, recordingID, objectKey, f, info.Size(), "video/mp4")
	if err != nil {
		return "", err
	}
	return "gdrive:" + fileID, nil
}

// LocationMissing is written when the local file is gone so the row leaves the
// unarchived queue and does not permanently block older-first batches.
const LocationMissing = "skipped:missing"

// ArchivePending uploads up to limit unarchived segments via index.
// Single-flight: concurrent cron/API calls serialize on one mutex.
// Missing local files are marked LocationMissing (skipped) so the queue advances.
// Upload then mark: mark is retried a few times after a successful upload. Once
// the Drive metadata is durable, the local file is removed. A local cleanup
// failure is reported separately because the remote archive is still safe.
func (s *Service) ArchivePending(ctx context.Context, index ArchiveIndex, limit int) (ArchiveStats, error) {
	var stats ArchiveStats
	if s == nil {
		return stats, fmt.Errorf("gdrive service is nil")
	}
	if index == nil {
		return stats, fmt.Errorf("archive index is nil")
	}
	if !s.archiveMu.TryLock() {
		return stats, ErrArchiveInProgress
	}
	defer s.archiveMu.Unlock()

	if !s.Connected(ctx) {
		return stats, fmt.Errorf("google drive is not connected")
	}
	if limit <= 0 {
		limit = DefaultArchiveBatchLimit
	}

	segs, err := index.ListUnarchived(ctx, limit)
	if err != nil {
		return stats, fmt.Errorf("list unarchived: %w", err)
	}

	for _, seg := range segs {
		if err := ctx.Err(); err != nil {
			return stats, err
		}

		abs, err := index.AbsolutePath(seg.Path)
		if err != nil {
			slog.Warn("gdrive archive: bad path", "id", seg.ID, "path", seg.Path, "err", err)
			stats.Failed++
			continue
		}
		if _, err := os.Stat(abs); err != nil {
			slog.Warn("gdrive archive: missing file", "id", seg.ID, "path", abs, "err", err)
			if !errors.Is(err, os.ErrNotExist) {
				stats.Failed++
				continue
			}
			if markErr := index.MarkArchived(ctx, seg.ID, LocationMissing); markErr != nil {
				slog.Warn("gdrive archive: mark missing failed", "id", seg.ID, "err", markErr)
				stats.Failed++
			} else {
				stats.Skipped++
			}
			continue
		}

		objectKey := filepath.ToSlash(seg.Path)
		archiveFile := s.archiveFile
		if archiveFile == nil {
			archiveFile = s.ArchiveLocalFile
		}
		loc, err := archiveFile(ctx, seg.ID, objectKey, abs)
		if err != nil {
			slog.Warn("gdrive archive: upload failed", "id", seg.ID, "path", objectKey, "err", err)
			stats.Failed++
			continue
		}
		if err := markArchivedWithRetry(ctx, index, seg.ID, loc); err != nil {
			slog.Warn("gdrive archive: mark failed after upload", "id", seg.ID, "location", loc, "err", err)
			stats.Failed++
			continue
		}
		stats.Uploaded++
		if err := index.DeleteLocal(ctx, seg.ID, seg.Path); err != nil {
			slog.Warn("gdrive archive: local cleanup failed", "id", seg.ID, "path", abs, "err", err)
			stats.DeleteFailed++
		} else {
			stats.Deleted++
		}
	}

	slog.Info("gdrive archive batch complete",
		"uploaded", stats.Uploaded,
		"deleted", stats.Deleted,
		"delete_failed", stats.DeleteFailed,
		"failed", stats.Failed,
		"skipped", stats.Skipped,
	)
	return stats, nil
}

func markArchivedWithRetry(ctx context.Context, index ArchiveIndex, id, location string) error {
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := index.MarkArchived(ctx, id, location); err != nil {
			last = err
			// Brief backoff without importing time-heavy helpers in loop of success path.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 100 * time.Millisecond):
			}
			continue
		}
		return nil
	}
	return last
}
