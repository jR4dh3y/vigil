package api

import (
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/shirou/gopsutil/v4/disk"
)

// GetSystemDisk returns disk usage for the recordings directory.
func (s *Server) GetSystemDisk(w http.ResponseWriter, r *http.Request) {
	if requireUser(w, r) == nil {
		return
	}
	info, err := s.diskInfo(r)
	if err != nil {
		slog.Error("disk usage", "err", err, "path", s.RecordingsDir)
		writeError(w, http.StatusInternalServerError, "disk usage unavailable", "internal")
		return
	}
	writeJSON(w, http.StatusOK, info)
}

// GetSystemStatus returns a system overview for the dashboard.
func (s *Server) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	if requireUser(w, r) == nil {
		return
	}

	info, err := s.diskInfo(r)
	if err != nil {
		slog.Error("disk usage", "err", err, "path", s.RecordingsDir)
		writeError(w, http.StatusInternalServerError, "disk usage unavailable", "internal")
		return
	}

	var total, enabled, online int
	if s.Camera != nil {
		list, err := s.Camera.List(r.Context())
		if err != nil {
			slog.Error("list cameras for status", "err", err)
			writeError(w, http.StatusInternalServerError, "internal error", "internal")
			return
		}
		total = len(list)
		for _, c := range list {
			if c.Enabled {
				enabled++
			}
			if c.Status == "online" {
				online++
			}
		}
	}

	retention := s.DefaultRetentionDays
	if s.Recording != nil {
		retention = s.Recording.RetentionDays()
	}
	// Prefer DB setting if present.
	if days, ok := s.readRetentionDays(r); ok {
		retention = days
	}

	healthStatus := SystemStatusHealthStatusOk
	if info.UsedPercent >= 90 {
		healthStatus = SystemStatusHealthStatusDegraded
	}

	status := SystemStatus{
		Version: s.Version,
		Commit:  s.Commit,
		Disk:    info,
		RetentionDays: retention,
	}
	status.Health.Status = healthStatus
	status.Cameras.Total = total
	status.Cameras.Enabled = enabled
	status.Cameras.Online = online
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) diskInfo(r *http.Request) (DiskInfo, error) {
	path := s.RecordingsDir
	if s.Recording != nil {
		if dir := s.Recording.RecordingsDir(); dir != "" {
			path = dir
		}
	}
	if s.Media != nil {
		if dir := s.Media.RecordingsDir(); dir != "" {
			path = dir
		}
	}
	if path == "" {
		path = "."
	}
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}
	usage, err := disk.UsageWithContext(r.Context(), path)
	if err != nil {
		return DiskInfo{}, err
	}
	return DiskInfo{
		Path:        path,
		TotalBytes:  int64(usage.Total),
		UsedBytes:   int64(usage.Used),
		FreeBytes:   int64(usage.Free),
		UsedPercent: float32(usage.UsedPercent),
	}, nil
}
