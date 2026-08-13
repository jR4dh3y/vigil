package api

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

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
	refreshMediaPaths := false

	if body.RetentionDays != nil {
		days := *body.RetentionDays
		if days < 1 {
			writeError(w, http.StatusBadRequest, "retentionDays must be at least 1", "validation")
			return
		}
		if err := s.Queries.UpsertSetting(r.Context(), store.UpsertSettingParams{
			Key:   settingKeyRetentionDays,
			Value: strconv.Itoa(days),
		}); err != nil {
			slog.Error("upsert retentionDays", "err", err)
			writeError(w, http.StatusInternalServerError, "internal error", "internal")
			return
		}
		if s.Recording != nil {
			s.Recording.SetRetentionDays(days)
		}
		if s.Media != nil {
			s.Media.SetRetentionDays(days)
			refreshMediaPaths = true
		}
	}

	if body.SiteName != nil {
		name := strings.TrimSpace(*body.SiteName)
		if err := s.Queries.UpsertSetting(r.Context(), store.UpsertSettingParams{
			Key:   settingKeySiteName,
			Value: name,
		}); err != nil {
			slog.Error("upsert siteName", "err", err)
			writeError(w, http.StatusInternalServerError, "internal error", "internal")
			return
		}
	}

	// Resolve next recording config from body + current values.
	if body.RecordingsDir != nil || body.RecordingEnabled != nil {
		current := s.loadSettings(r)
		nextDir := current.RecordingsDir
		nextEnabled := current.RecordingEnabled

		if body.RecordingsDir != nil {
			nextDir = strings.TrimSpace(*body.RecordingsDir)
			if nextDir != "" {
				nextDir = filepath.Clean(nextDir)
			}
		}
		if body.RecordingEnabled != nil {
			nextEnabled = *body.RecordingEnabled
		}
		if nextEnabled && nextDir == "" {
			writeError(w, http.StatusBadRequest, "recordingsDir is required when recording is enabled", "validation")
			return
		}

		if s.Media != nil {
			if err := s.Media.SetRecordingConfig(nextDir, nextEnabled); err != nil {
				slog.Error("set recording config", "err", err)
				writeError(w, http.StatusBadRequest, err.Error(), "validation")
				return
			}
			// Prefer cleaned path after mkdir.
			if abs := s.Media.RecordingsDir(); abs != "" {
				nextDir = abs
			}
			nextEnabled = s.Media.RecordingEnabled()
		}

		if err := s.Queries.UpsertSetting(r.Context(), store.UpsertSettingParams{
			Key:   settingKeyRecordingsDir,
			Value: nextDir,
		}); err != nil {
			slog.Error("upsert recordingsDir", "err", err)
			writeError(w, http.StatusInternalServerError, "internal error", "internal")
			return
		}
		enabledVal := "false"
		if nextEnabled {
			enabledVal = "true"
		}
		if err := s.Queries.UpsertSetting(r.Context(), store.UpsertSettingParams{
			Key:   settingKeyRecordingEnabled,
			Value: enabledVal,
		}); err != nil {
			slog.Error("upsert recordingEnabled", "err", err)
			writeError(w, http.StatusInternalServerError, "internal error", "internal")
			return
		}

		s.RecordingsDir = nextDir
		if s.Recording != nil {
			s.Recording.SetRecordingsDir(nextDir)
		}

		refreshMediaPaths = true
	}

	// Push retention and recording settings to every runtime MediaMTX path.
	// MediaMTX keeps path configuration in memory, so a database-only update
	// would leave its local deletion timer stale until the next restart.
	if refreshMediaPaths && s.Media != nil && s.Camera != nil {
		list, err := s.Camera.List(r.Context())
		if err != nil {
			slog.Warn("list cameras after recording settings change", "err", err)
		} else {
			s.Media.ReapplyCameraPaths(r.Context(), list)
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
