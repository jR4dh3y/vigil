package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/camera"
	"github.com/nvr/nvr/server/internal/media"
)

// listCamerasResponse matches GET /cameras OpenAPI response body.
type listCamerasResponse struct {
	Cameras []Camera `json:"cameras"`
}

// ListCameras returns all cameras for any authenticated user.
func (s *Server) ListCameras(w http.ResponseWriter, r *http.Request) {
	if auth.UserFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "unauthorized")
		return
	}
	if s.Camera == nil {
		writeError(w, http.StatusInternalServerError, "camera service unavailable", "internal")
		return
	}

	list, err := s.Camera.List(r.Context())
	if err != nil {
		slog.Error("list cameras", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	out := make([]Camera, 0, len(list))
	for _, c := range list {
		out = append(out, mapCamera(c))
	}
	writeJSON(w, http.StatusOK, listCamerasResponse{Cameras: out})
}

// CreateCamera creates a camera (admin or operator).
func (s *Server) CreateCamera(w http.ResponseWriter, r *http.Request) {
	if !requireOperator(w, r) {
		return
	}
	if s.Camera == nil {
		writeError(w, http.StatusInternalServerError, "camera service unavailable", "internal")
		return
	}

	var body CreateCameraRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	name := strings.TrimSpace(body.Name)
	host := strings.TrimSpace(body.Host)
	if name == "" || host == "" {
		writeError(w, http.StatusBadRequest, "name and host are required", "validation")
		return
	}

	in := camera.CreateInput{
		Name:    name,
		Host:    host,
		Enabled: true,
	}
	if body.Driver != nil {
		in.Driver = strings.TrimSpace(*body.Driver)
	}
	if body.Enabled != nil {
		in.Enabled = *body.Enabled
	}
	if body.Username != nil {
		in.Username = *body.Username
	}
	if body.Password != nil {
		in.Password = *body.Password
	}
	if body.LiveRtspUrl != nil {
		in.LiveRTSPURL = strings.TrimSpace(*body.LiveRtspUrl)
	}
	if body.RecordRtspUrl != nil {
		in.RecordRTSPURL = strings.TrimSpace(*body.RecordRtspUrl)
	}

	c, err := s.Camera.Create(r.Context(), in)
	if err != nil {
		slog.Error("create camera", "err", err)
		writeError(w, http.StatusBadRequest, err.Error(), "validation")
		return
	}
	s.syncMediaPath(r.Context(), c)
	writeJSON(w, http.StatusCreated, mapCamera(c))
}

// DiscoverCameras scans the local network for ONVIF cameras without credentials.
func (s *Server) DiscoverCameras(w http.ResponseWriter, r *http.Request) {
	if !requireOperator(w, r) {
		return
	}
	if s.Camera == nil {
		writeError(w, http.StatusInternalServerError, "camera service unavailable", "internal")
		return
	}

	list, err := s.Camera.Discover(r.Context())
	if err != nil {
		slog.Error("discover cameras", "err", err)
		writeError(w, http.StatusInternalServerError, "camera discovery failed", "internal")
		return
	}

	out := make([]DiscoveredCamera, 0, len(list))
	for _, discovered := range list {
		out = append(out, mapDiscoveredCamera(discovered))
	}
	writeJSON(w, http.StatusOK, DiscoverResult{Cameras: out})
}

// DiscoverCameraStreams authenticates to a selected ONVIF camera and returns
// the RTSP URLs reported by its media service.
func (s *Server) DiscoverCameraStreams(w http.ResponseWriter, r *http.Request) {
	if !requireOperator(w, r) {
		return
	}
	if s.Camera == nil {
		writeError(w, http.StatusInternalServerError, "camera service unavailable", "internal")
		return
	}

	var body DiscoverCameraStreamsRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	if strings.TrimSpace(body.Xaddr) == "" ||
		strings.TrimSpace(body.Username) == "" ||
		body.Password == nil ||
		*body.Password == "" {
		writeError(w, http.StatusBadRequest, "xaddr, username, and password are required", "validation")
		return
	}

	result, err := s.Camera.DiscoverStreams(r.Context(), camera.StreamDiscoveryInput{
		XAddr:    body.Xaddr,
		Username: body.Username,
		Password: *body.Password,
	})
	if err != nil {
		slog.Warn("discover camera streams", "err", err)
		writeError(w, http.StatusBadRequest, err.Error(), "discovery_failed")
		return
	}
	writeJSON(w, http.StatusOK, DiscoverCameraStreamsResult{
		LiveRtspUrl:   result.LiveRTSPURL,
		RecordRtspUrl: result.RecordRTSPURL,
	})
}

// GetCamera returns one camera for any authenticated user.
func (s *Server) GetCamera(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if auth.UserFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "unauthorized")
		return
	}
	if s.Camera == nil {
		writeError(w, http.StatusInternalServerError, "camera service unavailable", "internal")
		return
	}

	c, err := s.Camera.Get(r.Context(), id.String())
	if err != nil {
		if errors.Is(err, camera.ErrNotFound) {
			writeError(w, http.StatusNotFound, "camera not found", "not_found")
			return
		}
		slog.Error("get camera", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	writeJSON(w, http.StatusOK, mapCamera(c))
}

// UpdateCamera updates a camera (admin or operator).
func (s *Server) UpdateCamera(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if !requireOperator(w, r) {
		return
	}
	if s.Camera == nil {
		writeError(w, http.StatusInternalServerError, "camera service unavailable", "internal")
		return
	}

	var body UpdateCameraRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	in := camera.UpdateInput{
		Name:          body.Name,
		Driver:        body.Driver,
		Host:          body.Host,
		Username:      body.Username,
		Password:      body.Password,
		Enabled:       body.Enabled,
		LiveRTSPURL:   body.LiveRtspUrl,
		RecordRTSPURL: body.RecordRtspUrl,
	}

	c, err := s.Camera.Update(r.Context(), id.String(), in)
	if err != nil {
		if errors.Is(err, camera.ErrNotFound) {
			writeError(w, http.StatusNotFound, "camera not found", "not_found")
			return
		}
		slog.Error("update camera", "err", err)
		writeError(w, http.StatusBadRequest, err.Error(), "validation")
		return
	}
	s.syncMediaPath(r.Context(), c)
	writeJSON(w, http.StatusOK, mapCamera(c))
}

// DeleteCamera deletes a camera (admin or operator).
func (s *Server) DeleteCamera(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if !requireOperator(w, r) {
		return
	}
	if s.Camera == nil {
		writeError(w, http.StatusInternalServerError, "camera service unavailable", "internal")
		return
	}

	if err := s.Camera.Delete(r.Context(), id.String()); err != nil {
		if errors.Is(err, camera.ErrNotFound) {
			writeError(w, http.StatusNotFound, "camera not found", "not_found")
			return
		}
		slog.Error("delete camera", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	if s.Media != nil {
		if err := s.Media.DeletePath(r.Context(), id.String()); err != nil {
			slog.Warn("mediamtx delete path after camera delete", "camera_id", id.String(), "err", err)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// PostCameraLive mints live stream endpoints (any authenticated user).
func (s *Server) PostCameraLive(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if auth.UserFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "unauthorized")
		return
	}
	if s.Media == nil {
		writeError(w, http.StatusInternalServerError, "media service unavailable", "internal")
		return
	}

	live, err := s.Media.IssueLive(r.Context(), id.String())
	if err != nil {
		if errors.Is(err, camera.ErrNotFound) {
			writeError(w, http.StatusNotFound, "camera not found", "not_found")
			return
		}
		if errors.Is(err, media.ErrCameraDisabled) {
			writeError(w, http.StatusConflict, "camera is disabled", "camera_disabled")
			return
		}
		slog.Error("post camera live", "err", err, "camera_id", id.String())
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	writeJSON(w, http.StatusOK, LiveStream{
		CameraId:  live.CameraID,
		WhepUrl:   live.WHEPURL,
		HlsUrl:    live.HLSURL,
		Token:     live.Token,
		ExpiresAt: live.ExpiresAt,
	})
}

// GetCameraSnapshot returns a JPEG snapshot (any authenticated user).
func (s *Server) GetCameraSnapshot(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if auth.UserFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "unauthorized")
		return
	}
	if s.Media == nil {
		writeError(w, http.StatusInternalServerError, "media service unavailable", "internal")
		return
	}

	jpeg, err := s.Media.Snapshot(r.Context(), id.String())
	if err != nil {
		if errors.Is(err, camera.ErrNotFound) {
			writeError(w, http.StatusNotFound, "camera not found", "not_found")
			return
		}
		if errors.Is(err, media.ErrCameraDisabled) {
			writeError(w, http.StatusNotFound, "camera is disabled", "camera_disabled")
			return
		}
		if errors.Is(err, media.ErrNoLiveSource) || errors.Is(err, media.ErrSnapshotFailed) {
			writeError(w, http.StatusNotFound, "snapshot not available", "not_found")
			return
		}
		slog.Error("get camera snapshot", "err", err, "camera_id", id.String())
		writeError(w, http.StatusNotFound, "snapshot not available", "not_found")
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jpeg)
}

// ProbeCamera probes an RTSP URL without creating a camera (admin or operator).
func (s *Server) ProbeCamera(w http.ResponseWriter, r *http.Request) {
	if !requireOperator(w, r) {
		return
	}
	if s.Camera == nil {
		writeError(w, http.StatusInternalServerError, "camera service unavailable", "internal")
		return
	}

	var body ProbeCameraRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	if strings.TrimSpace(body.RtspUrl) == "" {
		writeError(w, http.StatusBadRequest, "rtspUrl is required", "validation")
		return
	}

	user, pass := "", ""
	if body.Username != nil {
		user = *body.Username
	}
	if body.Password != nil {
		pass = *body.Password
	}

	result, err := s.Camera.Probe(r.Context(), body.RtspUrl, user, pass)
	if err != nil {
		slog.Error("probe camera", "err", err)
		writeError(w, http.StatusBadRequest, err.Error(), "validation")
		return
	}
	writeJSON(w, http.StatusOK, mapProbeResult(result))
}

// syncMediaPath ensures or removes the MediaMTX path based on camera enabled state.
// Failures are logged only (best-effort).
func (s *Server) syncMediaPath(ctx context.Context, c camera.Camera) {
	if s.Media == nil {
		return
	}
	if c.Enabled {
		if err := s.Media.EnsurePathForCamera(ctx, c); err != nil {
			slog.Warn("mediamtx ensure path", "camera_id", c.ID, "err", err)
		}
		return
	}
	if err := s.Media.DeletePath(ctx, c.ID); err != nil {
		slog.Warn("mediamtx delete path", "camera_id", c.ID, "err", err)
	}
}

func mapDiscoveredCamera(c camera.DiscoveredCamera) DiscoveredCamera {
	return DiscoveredCamera{
		Id:    c.ID,
		Name:  c.Name,
		Host:  c.Host,
		Xaddr: c.XAddr,
	}
}

func mapCamera(c camera.Camera) Camera {
	id, err := uuid.Parse(c.ID)
	if err != nil {
		// IDs are always UUIDs from Create; fall back to zero UUID on parse failure.
		id = uuid.Nil
	}

	profiles := make([]StreamProfile, 0, len(c.StreamProfiles))
	for _, p := range c.StreamProfiles {
		profiles = append(profiles, StreamProfile{
			Id:      p.ID,
			Role:    StreamProfileRole(p.Role),
			RtspUrl: p.RTSPURL,
			Codec:   p.Codec,
			Width:   p.Width,
			Height:  p.Height,
		})
	}

	status := CameraStatus(c.Status)
	if !status.Valid() {
		status = Unknown
	}

	return Camera{
		Id:             id,
		Name:           c.Name,
		Driver:         c.Driver,
		Enabled:        c.Enabled,
		Host:           c.Host,
		Status:         status,
		StreamProfiles: profiles,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func mapProbeResult(r camera.ProbeResult) ProbeResult {
	out := ProbeResult{
		Reachable: r.Reachable,
		H265:      r.H265,
	}
	if r.Codec != "" {
		c := r.Codec
		out.Codec = &c
	}
	if r.Width > 0 {
		w := r.Width
		out.Width = &w
	}
	if r.Height > 0 {
		h := r.Height
		out.Height = &h
	}
	if r.Error != "" {
		e := r.Error
		out.Error = &e
	}
	return out
}
