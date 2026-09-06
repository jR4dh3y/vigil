package storage

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/nvr/nvr/server/internal/recording"
	"github.com/nvr/nvr/server/internal/storage/gdrive"
)

var (
	// ErrGDriveNotConfigured indicates that no Google Drive service is available.
	ErrGDriveNotConfigured = errors.New("google drive is not configured")
	// ErrRecordingNotConfigured indicates that no recording service is available.
	ErrRecordingNotConfigured = errors.New("recording service is not configured")
	// ErrGDriveNotConnected indicates that the configured Drive service has no usable credentials.
	ErrGDriveNotConnected = errors.New("google drive is not connected")
)

// RecordingArchiveIndex adapts recording.Service to the archive storage seam.
// Keeping the adapter here avoids duplicating cross-domain mapping in API and jobs.
type RecordingArchiveIndex struct {
	Recording *recording.Service
}

// PrepareGDriveArchive validates the shared API/job preconditions and returns
// the recording index used by a Drive archive pass.
func PrepareGDriveArchive(
	ctx context.Context,
	drive *gdrive.Service,
	recordings *recording.Service,
) (RecordingArchiveIndex, error) {
	if drive == nil {
		return RecordingArchiveIndex{}, ErrGDriveNotConfigured
	}
	if recordings == nil {
		return RecordingArchiveIndex{}, ErrRecordingNotConfigured
	}
	if !drive.Connected(ctx) {
		return RecordingArchiveIndex{}, ErrGDriveNotConnected
	}
	stats, err := recordings.ReconcileDisk(ctx, 5*time.Second)
	if err != nil {
		return RecordingArchiveIndex{}, err
	}
	if stats.Indexed > 0 || stats.Failed > 0 {
		slog.Info("recording pre-archive reconciliation complete",
			"indexed", stats.Indexed,
			"existing", stats.Existing,
			"skipped", stats.Skipped,
			"failed", stats.Failed,
		)
	}
	return RecordingArchiveIndex{Recording: recordings}, nil
}

func (a RecordingArchiveIndex) ListUnarchived(ctx context.Context, limit int) ([]gdrive.ArchiveSegment, error) {
	segs, err := a.Recording.ListUnarchived(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]gdrive.ArchiveSegment, 0, len(segs))
	for _, segment := range segs {
		out = append(out, gdrive.ArchiveSegment{
			ID:        segment.ID,
			Path:      segment.Path,
			CameraID:  segment.CameraID,
			StartedAt: segment.StartedAt,
		})
	}
	return out, nil
}

func (a RecordingArchiveIndex) AbsolutePath(rel string) (string, error) {
	return a.Recording.AbsolutePath(rel)
}

func (a RecordingArchiveIndex) MarkArchived(ctx context.Context, id, location string) error {
	err := a.Recording.MarkArchived(ctx, id, location)
	if errors.Is(err, recording.ErrAlreadyArchived) {
		return gdrive.ErrAlreadyArchived
	}
	return err
}

func (a RecordingArchiveIndex) DeleteLocal(ctx context.Context, id, path, absolutePath string) error {
	return a.Recording.DeleteLocalAt(ctx, id, path, absolutePath)
}

func (a RecordingArchiveIndex) LockArchive(ctx context.Context) (func(), error) {
	if a.Recording == nil {
		return func() {}, nil
	}
	return a.Recording.LockArchive(ctx)
}
