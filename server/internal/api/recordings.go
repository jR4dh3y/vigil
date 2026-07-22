package api

import (
	"errors"
	"log/slog"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/camera"
	"github.com/nvr/nvr/server/internal/recording"
)

// ListCameraRecordings returns recording segments and coverage for a time range.
// Any authenticated user; 404 if the camera does not exist.
func (s *Server) ListCameraRecordings(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params ListCameraRecordingsParams) {
	if auth.UserFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "unauthorized")
		return
	}
	if s.Recording == nil || s.Camera == nil {
		writeError(w, http.StatusInternalServerError, "recording service unavailable", "internal")
		return
	}

	cameraID := id.String()
	if _, err := s.Camera.Get(r.Context(), cameraID); err != nil {
		if errors.Is(err, camera.ErrNotFound) {
			writeError(w, http.StatusNotFound, "camera not found", "not_found")
			return
		}
		slog.Error("list recordings get camera", "err", err, "camera_id", cameraID)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	result, err := s.Recording.List(r.Context(), cameraID, params.From, params.To)
	if err != nil {
		slog.Error("list recordings", "err", err, "camera_id", cameraID)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	writeJSON(w, http.StatusOK, mapRecordingList(result))
}

// PostCameraPlayback mints a short-lived playback session for recorded video.
// Any authenticated user; 404 if the camera does not exist.
func (s *Server) PostCameraPlayback(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if auth.UserFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "unauthorized")
		return
	}
	if s.Media == nil {
		writeError(w, http.StatusInternalServerError, "media service unavailable", "internal")
		return
	}

	var body PlaybackRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	if body.Start.IsZero() {
		writeError(w, http.StatusBadRequest, "start is required", "validation")
		return
	}

	durationSec := 60.0
	if body.DurationSec != nil && *body.DurationSec > 0 {
		durationSec = float64(*body.DurationSec)
	}

	session, err := s.Media.IssuePlayback(r.Context(), id.String(), body.Start, durationSec)
	if err != nil {
		if errors.Is(err, camera.ErrNotFound) {
			writeError(w, http.StatusNotFound, "camera not found", "not_found")
			return
		}
		slog.Error("post camera playback", "err", err, "camera_id", id.String())
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	writeJSON(w, http.StatusOK, PlaybackSession{
		CameraId:    session.CameraID,
		PlaybackUrl: session.PlaybackURL,
		Token:       session.Token,
		ExpiresAt:   session.ExpiresAt,
	})
}

func mapRecordingList(result recording.ListResult) RecordingList {
	out := RecordingList{
		Recordings: make([]RecordingSegment, 0, len(result.Segments)),
		Coverage:   make([]CoverageBar, 0, len(result.Coverage)),
	}
	for _, seg := range result.Segments {
		item := RecordingSegment{
			Id:          seg.ID,
			CameraId:    seg.CameraID,
			StartedAt:   seg.StartedAt,
			DurationSec: float32(seg.DurationSec),
			SizeBytes:   seg.SizeBytes,
			Path:        seg.Path,
			Codec:       seg.Codec,
		}
		// ThumbnailUrl left nil until thumbnail serving is implemented.
		out.Recordings = append(out.Recordings, item)
	}
	for _, bar := range result.Coverage {
		out.Coverage = append(out.Coverage, CoverageBar{
			Start: bar.Start,
			End:   bar.End,
		})
	}
	return out
}
