package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nvr/nvr/server/internal/camera"
	"github.com/nvr/nvr/server/internal/media"
	"github.com/nvr/nvr/server/internal/store"
)

const (
	settingKeyRetentionDays    = "retentionDays"
	settingKeySiteName         = "siteName"
	settingKeyRecordingsDir    = "recordingsDir"
	settingKeyRecordingEnabled = "recordingEnabled"
	defaultSiteName            = "NVR"
)

// GetSettings returns site settings for any authenticated user.
func (s *Server) GetSettings(w http.ResponseWriter, r *http.Request) {
	if requireUser(w, r) == nil {
		return
	}
	writeJSON(w, http.StatusOK, s.loadSettings(r))
}

// PatchSettings updates settings (admin only).
func (s *Server) PatchSettings(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}

	var body PatchSettingsRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	// Settings updates include a runtime MediaMTX change and must not race one
	// another while reading current values, applying paths, or committing DB
	// state. The lock is intentionally held through the path synchronization.
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()

	current := s.loadSettings(r)
	nextRetention := current.RetentionDays
	if body.RetentionDays != nil {
		nextRetention = *body.RetentionDays
		if nextRetention < 1 {
			writeError(w, http.StatusBadRequest, "retentionDays must be at least 1", "validation")
			return
		}
	}

	nextSiteName := current.SiteName
	if body.SiteName != nil {
		nextSiteName = strings.TrimSpace(*body.SiteName)
	}

	recordingConfigChanged := body.RecordingsDir != nil || body.RecordingEnabled != nil
	nextDir := current.RecordingsDir
	nextEnabled := current.RecordingEnabled
	if body.RecordingsDir != nil {
		nextDir = *body.RecordingsDir
	}
	if body.RecordingEnabled != nil {
		nextEnabled = *body.RecordingEnabled
	}
	if recordingConfigChanged {
		var err error
		nextDir, nextEnabled, err = media.NormalizeRecordingConfig(nextDir, nextEnabled)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error(), "validation")
			return
		}
	}

	retentionChanged := body.RetentionDays != nil
	refreshMediaPaths := s.Media != nil && (retentionChanged || recordingConfigChanged)
	var restoreRuntime func()
	if refreshMediaPaths {
		oldDir := s.Media.RecordingsDir()
		oldEnabled := s.Media.RecordingEnabled()
		oldRetention := s.Media.RetentionDays()

		var cameraList []camera.Camera
		if s.Camera != nil {
			var err error
			cameraList, err = s.Camera.List(r.Context())
			if err != nil {
				slog.Warn("list cameras before recording settings change", "err", err)
				writeError(w, http.StatusServiceUnavailable, "could not synchronize camera recording paths", "mediamtx_sync_failed")
				return
			}
		}

		restoreRuntime = func() {
			// Rollback must still run when the request was cancelled after runtime
			// state changed. Keep it bounded but independent of r.Context().
			restoreCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			if recordingConfigChanged {
				if err := s.Media.SetRecordingConfig(oldDir, oldEnabled); err != nil {
					slog.Warn("restore recording config after settings failure", "err", err)
				}
			}
			if retentionChanged {
				s.Media.SetRetentionDays(oldRetention)
			}
			if s.Camera != nil {
				if err := s.Media.ReapplyCameraPaths(restoreCtx, cameraList); err != nil {
					slog.Warn("restore mediamtx paths after settings failure", "err", err)
				}
			}
		}

		if recordingConfigChanged {
			if err := s.Media.SetRecordingConfig(nextDir, nextEnabled); err != nil {
				writeError(w, http.StatusBadRequest, err.Error(), "validation")
				return
			}
			nextDir = s.Media.RecordingsDir()
			nextEnabled = s.Media.RecordingEnabled()
		}
		if retentionChanged {
			s.Media.SetRetentionDays(nextRetention)
		}
		if s.Camera != nil {
			if err := s.Media.ReapplyCameraPaths(r.Context(), cameraList); err != nil {
				restoreRuntime()
				writeError(w, http.StatusServiceUnavailable, "could not synchronize camera recording paths", "mediamtx_sync_failed")
				return
			}
		}
	}

	// Commit all requested settings together. Runtime changes are rolled back
	// best-effort if the database transaction fails.
	err := store.InTransaction(r.Context(), s.Queries, func(q *store.Queries) error {
		if retentionChanged {
			if err := q.UpsertSetting(r.Context(), store.UpsertSettingParams{
				Key:   settingKeyRetentionDays,
				Value: strconv.Itoa(nextRetention),
			}); err != nil {
				return fmt.Errorf("upsert retentionDays: %w", err)
			}
		}
		if body.SiteName != nil {
			if err := q.UpsertSetting(r.Context(), store.UpsertSettingParams{
				Key:   settingKeySiteName,
				Value: nextSiteName,
			}); err != nil {
				return fmt.Errorf("upsert siteName: %w", err)
			}
		}
		if recordingConfigChanged {
			if err := q.UpsertSetting(r.Context(), store.UpsertSettingParams{
				Key:   settingKeyRecordingsDir,
				Value: nextDir,
			}); err != nil {
				return fmt.Errorf("upsert recordingsDir: %w", err)
			}
			enabledVal := "false"
			if nextEnabled {
				enabledVal = "true"
			}
			if err := q.UpsertSetting(r.Context(), store.UpsertSettingParams{
				Key:   settingKeyRecordingEnabled,
				Value: enabledVal,
			}); err != nil {
				return fmt.Errorf("upsert recordingEnabled: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		if restoreRuntime != nil {
			restoreRuntime()
		}
		slog.Error("commit settings", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	if retentionChanged && s.Recording != nil {
		s.Recording.SetRetentionDays(nextRetention)
	}
	if recordingConfigChanged {
		s.RecordingsDir = nextDir
		if s.Recording != nil {
			s.Recording.SetRecordingsDir(nextDir)
		}
	}

	writeJSON(w, http.StatusOK, s.loadSettings(r))
}

func (s *Server) loadSettings(r *http.Request) Settings {
	settings := Settings{
		RetentionDays:    s.DefaultRetentionDays,
		SiteName:         defaultSiteName,
		RecordingsDir:    s.RecordingsDir,
		RecordingEnabled: false,
	}
	if s.Recording != nil {
		settings.RetentionDays = s.Recording.RetentionDays()
		if dir := s.Recording.RecordingsDir(); dir != "" {
			settings.RecordingsDir = dir
		}
	}
	// DB values for site metadata; recording path/flag prefer live media runtime.
	if days, ok := s.readRetentionDays(r); ok {
		settings.RetentionDays = days
	}
	if name, ok := s.readSetting(r, settingKeySiteName); ok {
		settings.SiteName = name
	}
	if dir, ok := s.readSetting(r, settingKeyRecordingsDir); ok {
		settings.RecordingsDir = dir
	}
	if en, ok := s.readSetting(r, settingKeyRecordingEnabled); ok {
		settings.RecordingEnabled = parseBoolSetting(en, settings.RecordingEnabled)
	}
	if s.Media != nil {
		if dir := s.Media.RecordingsDir(); dir != "" {
			settings.RecordingsDir = dir
		}
		settings.RecordingEnabled = s.Media.RecordingEnabled()
	}
	return settings
}

func (s *Server) readRetentionDays(r *http.Request) (int, bool) {
	val, ok := s.readSetting(r, settingKeyRetentionDays)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil || n < 1 {
		return 0, false
	}
	return n, true
}

func (s *Server) readSetting(r *http.Request, key string) (string, bool) {
	if s.Queries == nil {
		return "", false
	}
	row, err := s.Queries.GetSetting(r.Context(), key)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			slog.Warn("get setting", "key", key, "err", err)
		}
		return "", false
	}
	return row.Value, true
}

func parseBoolSetting(val string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
