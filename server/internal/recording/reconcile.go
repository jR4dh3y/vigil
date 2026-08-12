package recording

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const defaultRecordingSettleAge = 5 * time.Second

// ReconcileStats summarizes a recordings-directory reconciliation pass.
type ReconcileStats struct {
	Discovered int
	Indexed    int
	Existing   int
	Skipped    int
	Failed     int
}

// ReconcileDisk indexes completed MP4 files that were missed by the MediaMTX
// completion hook. Files modified within settleAge are skipped because they may
// still be actively written. Repeated passes are safe because path is unique.
func (s *Service) ReconcileDisk(ctx context.Context, settleAge time.Duration) (ReconcileStats, error) {
	var stats ReconcileStats
	s.reconcileMu.Lock()
	defer s.reconcileMu.Unlock()

	root := strings.TrimSpace(s.recordingsDir)
	if root == "" {
		return stats, fmt.Errorf("recordings dir not configured")
	}
	if settleAge <= 0 {
		settleAge = defaultRecordingSettleAge
	}
	cutoff := time.Now().Add(-settleAge)

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".mp4") {
			return nil
		}

		stats.Discovered++
		info, err := entry.Info()
		if err != nil {
			stats.Failed++
			slog.Warn("recording reconcile: stat failed", "path", path, "err", err)
			return nil
		}
		if info.ModTime().After(cutoff) {
			stats.Skipped++
			return nil
		}

		cameraID, ok := cameraIDFromFilePath(path)
		if !ok {
			stats.Skipped++
			return nil
		}
		if _, err := s.q.GetCamera(ctx, cameraID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				stats.Skipped++
				return nil
			}
			stats.Failed++
			slog.Warn("recording reconcile: camera lookup failed", "camera_id", cameraID, "err", err)
			return nil
		}
		startedAt, ok := startedAtFromFileName(path)
		if !ok {
			stats.Skipped++
			return nil
		}

		rel := s.relativizePath(path)
		if _, err := s.q.GetRecordingByPath(ctx, rel); err == nil {
			stats.Existing++
			return nil
		} else if !errors.Is(err, sql.ErrNoRows) {
			stats.Failed++
			slog.Warn("recording reconcile: lookup failed", "path", rel, "err", err)
			return nil
		}

		duration, codec := probeRecording(ctx, path)
		if duration <= 0 {
			duration = 60
		}
		if _, err := s.IndexSegment(ctx, cameraID, path, startedAt, duration, info.Size(), codec); err != nil {
			stats.Failed++
			slog.Warn("recording reconcile: index failed", "path", path, "camera_id", cameraID, "err", err)
			return nil
		}
		stats.Indexed++
		return nil
	})
	if err != nil {
		return stats, fmt.Errorf("walk recordings dir: %w", err)
	}
	return stats, nil
}

func probeRecording(ctx context.Context, path string) (float64, string) {
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		probeCtx,
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name:format=duration",
		"-of", "json",
		path,
	)
	raw, err := cmd.Output()
	if err != nil {
		return 0, ""
	}
	var result struct {
		Streams []struct {
			CodecName string `json:"codec_name"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return 0, ""
	}
	duration, _ := strconv.ParseFloat(result.Format.Duration, 64)
	codec := ""
	if len(result.Streams) > 0 {
		codec = result.Streams[0].CodecName
	}
	return duration, codec
}
