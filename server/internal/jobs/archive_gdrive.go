package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nvr/nvr/server/internal/event"
	"github.com/nvr/nvr/server/internal/storage"
	"github.com/nvr/nvr/server/internal/storage/gdrive"
)

// DefaultGDriveArchiveLimit is the max segments archived per cron/API run when unspecified.
// This limit is aligned with the API's maximum validation (see api/storage_gdrive.go).
const DefaultGDriveArchiveLimit = 50

// RunGDriveArchive archives up to limit unarchived recordings to Google Drive.
// Safe to call from the API and from the midnight cron job.
func (s *Scheduler) RunGDriveArchive(ctx context.Context, limit int) (gdrive.ArchiveStats, error) {
	var zero gdrive.ArchiveStats
	index, err := storage.PrepareGDriveArchive(ctx, s.cfg.GDrive, s.cfg.Recording)
	if err != nil {
		return zero, err
	}
	if limit <= 0 {
		limit = DefaultGDriveArchiveLimit
	}
	return s.cfg.GDrive.ArchivePending(ctx, index, limit)
}

func (s *Scheduler) archiveToGDrive(ctx context.Context) {
	if s.cfg.GDrive == nil || s.cfg.Recording == nil {
		return
	}
	if !s.cfg.GDrive.Connected(ctx) {
		return
	}

	stats, err := s.RunGDriveArchive(ctx, DefaultGDriveArchiveLimit)
	if err != nil {
		slog.Warn("gdrive archive job failed", "err", err)
		return
	}
	slog.Info("gdrive archive job finished",
		"uploaded", stats.Uploaded,
		"failed", stats.Failed,
		"skipped", stats.Skipped,
	)

	if s.cfg.Events == nil {
		return
	}
	if stats.Uploaded == 0 && stats.Failed == 0 && stats.Skipped == 0 {
		return
	}
	severity := event.SeverityInfo
	title := "Google Drive archive complete"
	if stats.Failed > 0 {
		severity = event.SeverityWarning
		title = "Google Drive archive completed with errors"
	}
	if _, err := s.cfg.Events.Emit(ctx, event.EventInput{
		Type:      event.TypeArchiveComplete,
		Severity:  severity,
		Title:     title,
		Message:   fmt.Sprintf("Uploaded %d, failed %d, skipped %d", stats.Uploaded, stats.Failed, stats.Skipped),
		StartedAt: time.Now().UTC(),
		Metadata: map[string]any{
			"uploaded": stats.Uploaded,
			"failed":   stats.Failed,
			"skipped":  stats.Skipped,
		},
	}); err != nil {
		slog.Warn("gdrive archive: emit event failed", "err", err)
	}
}
