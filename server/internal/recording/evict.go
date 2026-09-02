package recording

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
)

// LocationExpired marks rows whose local file was removed before a Drive
// archive could happen (RAM-staging dwell limit or overflow eviction). Like
// LocationMissing it removes the row from the unarchived queue, and
// PruneArchived deletes these rows after the retention window.
const LocationExpired = "skipped:expired"

// maxEvictionBatch bounds one EnforceLocalLimits pass. RAM-backed staging
// volumes hold far fewer segments than this under normal operation.
const maxEvictionBatch = 1000

// LocalLimits bounds how long unarchived segments may occupy RecordingsDir.
// Both limits are opt-in; zero values disable them so default deployments are
// unaffected.
type LocalLimits struct {
	// MaxDwell evicts unarchived segments whose end time is older than
	// now-MaxDwell, oldest first. 0 disables dwell enforcement.
	MaxDwell time.Duration
	// MaxUsedPercent evicts oldest-first while the recordings volume is at or
	// above this used-percent. <= 0 disables overflow enforcement.
	MaxUsedPercent float64
}

// EvictStats summarizes one EnforceLocalLimits pass.
type EvictStats struct {
	// Examined is the number of unarchived rows considered.
	Examined int
	// Removed is the number of local files deleted.
	Removed int
	// FreedBytes approximates space reclaimed by removals.
	FreedBytes int64
	// BusySkipped counts files still within the settle age (possibly being
	// written by MediaMTX).
	BusySkipped int
	// Failed counts rows that could not be processed this pass.
	Failed int
}

// EnforceLocalLimits evicts local segment files that violate the supplied
// limits. Dwell violations are removed first (oldest overdue first), then
// overflow eviction removes oldest-first until volume usage drops below the
// threshold. Removed rows are marked LocationExpired so they leave the
// archive queue and stay visible (as non-preserved) on the timeline.
//
// EnforceLocalLimits serializes on LockArchive so it never races with Drive
// archive passes. Expired state is durably committed to SQLite before any local
// file is deleted, and failures are recorded rather than suppressed.
func (s *Service) EnforceLocalLimits(ctx context.Context, limits LocalLimits) (EvictStats, error) {
	var stats EvictStats
	if s == nil || s.q == nil {
		return stats, fmt.Errorf("recording service is not initialized")
	}
	if limits.MaxDwell <= 0 && limits.MaxUsedPercent <= 0 {
		return stats, nil
	}
	root := strings.TrimSpace(s.recordingsDir)
	if root == "" {
		return stats, fmt.Errorf("recordings dir not configured")
	}

	unlock, err := s.LockArchive(ctx)
	if err != nil {
		return stats, err
	}
	defer unlock()

	rows, err := s.q.ListUnarchivedRecordings(ctx, maxEvictionBatch)
	if err != nil {
		return stats, fmt.Errorf("list unarchived for eviction: %w", err)
	}
	stats.Examined = len(rows)

	segs := make([]Segment, 0, len(rows))
	for _, row := range rows {
		seg, err := toSegment(row)
		if err != nil {
			stats.Failed++
			continue
		}
		segs = append(segs, seg)
	}

	evicted := make(map[string]bool)

	// Pass 1: dwell expiry. Rows arrive oldest-first from the index query.
	if limits.MaxDwell > 0 {
		cutoff := time.Now().UTC().Add(-limits.MaxDwell)
		for _, seg := range segs {
			if ctx.Err() != nil {
				return stats, ctx.Err()
			}
			if !segmentEndedBefore(seg, cutoff) {
				continue
			}
			s.applyEviction(ctx, seg, &stats)
			evicted[seg.ID] = true
		}
	}

	// Pass 2: overflow eviction until usage drops below the threshold.
	if limits.MaxUsedPercent > 0 {
		// First retry cleanup of any already-archived or expired local files
		// before evicting fresh recordings.
		if _, err := s.CleanupArchivedLocals(ctx, defaultRecordingSettleAge); err != nil {
			slog.Warn("pre-overflow cleanup failed", "err", err)
		}
		for _, seg := range segs {
			if evicted[seg.ID] {
				continue
			}
			if ctx.Err() != nil {
				return stats, ctx.Err()
			}
			used, err := s.volumeUsage(ctx, root)
			if err != nil {
				return stats, fmt.Errorf("measure recordings volume: %w", err)
			}
			if used < limits.MaxUsedPercent {
				break
			}
			s.applyEviction(ctx, seg, &stats)
			evicted[seg.ID] = true
		}
	}
	return stats, nil
}

// applyEviction removes one segment's local file and marks the row expired,
// updating stats. The expired state is committed to SQLite BEFORE any local
// file is deleted so durable state is never lost, and update failures are
// tracked in stats.Failed instead of being silently ignored.
func (s *Service) applyEviction(ctx context.Context, seg Segment, stats *EvictStats) {
	abs, err := s.AbsolutePath(seg.Path)
	if err != nil {
		slog.Warn("resolve eviction path: invalid path", "id", seg.ID, "path", seg.Path, "err", err)
		stats.Failed++
		return
	}
	info, err := os.Lstat(abs)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			slog.Warn("stat eviction path failed", "id", seg.ID, "path", abs, "err", err)
			stats.Failed++
			return
		}
		// File already gone. Mark LocationExpired so the row leaves the queue.
		if markErr := s.MarkArchived(ctx, seg.ID, LocationExpired); markErr != nil {
			if !errors.Is(markErr, ErrAlreadyArchived) {
				slog.Warn("mark expired recording failed", "id", seg.ID, "err", markErr)
				stats.Failed++
			}
		}
		return
	}
	if !info.Mode().IsRegular() {
		slog.Warn("eviction candidate is not a regular file", "id", seg.ID, "path", abs)
		stats.Failed++
		return
	}
	if time.Since(info.ModTime()) < defaultRecordingSettleAge {
		stats.BusySkipped++
		return
	}

	// Commit expired state BEFORE deleting the file from disk.
	if markErr := s.MarkArchived(ctx, seg.ID, LocationExpired); markErr != nil {
		if errors.Is(markErr, ErrAlreadyArchived) {
			// Row was already archived or expired; leave local file alone.
			return
		}
		slog.Warn("mark evicted recording failed", "id", seg.ID, "err", markErr)
		stats.Failed++
		return
	}

	// Local state is durably marked LocationExpired; now reclaim space.
	size := info.Size()
	if err := os.Remove(abs); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Warn("remove evicted recording file failed", "id", seg.ID, "path", abs, "err", err)
		stats.Failed++
		return
	}
	stats.Removed++
	stats.FreedBytes += size
}

// volumeUsage reports used-percent of the recordings volume. usageFn overrides
// this in tests; production uses gopsutil statfs.
func (s *Service) volumeUsage(ctx context.Context, root string) (float64, error) {
	if s.usageFn != nil {
		return s.usageFn(ctx)
	}
	usage, err := disk.UsageWithContext(ctx, root)
	if err != nil {
		return 0, fmt.Errorf("statfs %s: %w", root, err)
	}
	return usage.UsedPercent, nil
}

// segmentEndedBefore reports whether seg's end time is before cutoff. Rows
// without a usable duration fall back to their start time.
func segmentEndedBefore(seg Segment, cutoff time.Time) bool {
	end := seg.StartedAt.Add(time.Duration(max(0, seg.DurationSec) * float64(time.Second)))
	return end.Before(cutoff)
}
