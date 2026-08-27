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

	// Pass 1: dwell expiry. Rows arrive oldest-first from the index query.
	start := 0
	if limits.MaxDwell > 0 {
		cutoff := time.Now().UTC().Add(-limits.MaxDwell)
		for i, seg := range segs {
			if ctx.Err() != nil {
				return stats, ctx.Err()
			}
			if !segmentEndedBefore(seg, cutoff) {
				continue
			}
			s.applyEviction(ctx, seg, &stats)
			start = i + 1
		}
	}

	// Pass 2: overflow eviction until usage drops below the threshold.
	// Rows already handled by pass 1 keep their order, so scanning resumes
	// right after the dwell prefix instead of retrying them.
	if limits.MaxUsedPercent > 0 {
		for _, seg := range segs[start:] {
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
		}
	}
	return stats, nil
}

// applyEviction removes one segment's local file and marks the row expired,
// updating stats. Failures are logged and counted, never fatal to the pass.
func (s *Service) applyEviction(ctx context.Context, seg Segment, stats *EvictStats) {
	removed, freed, busy, err := s.evictOne(ctx, seg)
	switch {
	case err != nil:
		slog.Warn("local eviction failed", "id", seg.ID, "path", seg.Path, "err", err)
		stats.Failed++
	case busy:
		stats.BusySkipped++
	default:
		if removed {
			stats.Removed++
			stats.FreedBytes += freed
		}
		if markErr := s.MarkArchived(ctx, seg.ID, LocationExpired); markErr != nil {
			slog.Warn("mark evicted recording failed", "id", seg.ID, "err", markErr)
		}
	}
}

// evictOne removes the local file for seg. It reports whether the file was
// removed, the bytes freed, and whether the file was skipped because it may
// still be written (settle age). A missing file is treated as success with
// removed=false so marking can proceed.
func (s *Service) evictOne(ctx context.Context, seg Segment) (removed bool, freed int64, busy bool, err error) {
	abs, err := s.AbsolutePath(seg.Path)
	if err != nil {
		return false, 0, false, fmt.Errorf("resolve eviction path: %w", err)
	}
	// Lstat deliberately refuses symlinks, mirroring DeleteLocalAt.
	info, err := os.Lstat(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, 0, false, nil
		}
		return false, 0, false, fmt.Errorf("stat %s: %w", abs, err)
	}
	if !info.Mode().IsRegular() {
		return false, 0, false, fmt.Errorf("%s is not a regular file", abs)
	}
	if time.Since(info.ModTime()) < defaultRecordingSettleAge {
		return false, 0, true, nil
	}
	if err := os.Remove(abs); err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, 0, false, fmt.Errorf("remove %s: %w", abs, err)
	}
	return true, info.Size(), false, nil
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
