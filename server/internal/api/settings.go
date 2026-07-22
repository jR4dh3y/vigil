package api

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/nvr/nvr/server/internal/store"
)

const (
	settingKeyRetentionDays = "retentionDays"
	settingKeySiteName      = "siteName"
	defaultSiteName         = "NVR"
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

	writeJSON(w, http.StatusOK, s.loadSettings(r))
}

func (s *Server) loadSettings(r *http.Request) Settings {
	settings := Settings{
		RetentionDays: s.DefaultRetentionDays,
		SiteName:      defaultSiteName,
	}
	if s.Recording != nil {
		settings.RetentionDays = s.Recording.RetentionDays()
	}
	if days, ok := s.readRetentionDays(r); ok {
		settings.RetentionDays = days
	}
	if name, ok := s.readSetting(r, settingKeySiteName); ok {
		settings.SiteName = name
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
