package storage

import (
	"context"

	"github.com/nvr/nvr/server/internal/recording"
	"github.com/nvr/nvr/server/internal/storage/gdrive"
)

// RecordingArchiveIndex adapts recording.Service to the archive storage seam.
// Keeping the adapter here avoids duplicating cross-domain mapping in API and jobs.
type RecordingArchiveIndex struct {
	Recording *recording.Service
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
	return a.Recording.MarkArchived(ctx, id, location)
}
