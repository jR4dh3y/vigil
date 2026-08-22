package api

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/camera"
	"github.com/nvr/nvr/server/internal/media"
	"github.com/nvr/nvr/server/internal/recording"
	"github.com/nvr/nvr/server/internal/storage/gdrive"
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

// ListRecordingDays returns recording storage availability grouped in the
// browser's local calendar days across all enabled cameras.
func (s *Server) ListRecordingDays(w http.ResponseWriter, r *http.Request, params ListRecordingDaysParams) {
	if auth.UserFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "unauthorized")
		return
	}
	if s.Recording == nil || s.Camera == nil {
		writeError(w, http.StatusInternalServerError, "recording service unavailable", "internal")
		return
	}

	location, err := time.LoadLocation(strings.TrimSpace(params.TimeZone))
	if err != nil {
		writeError(w, http.StatusBadRequest, "timeZone must be a valid IANA time zone", "validation")
		return
	}
	cameras, err := s.Camera.List(r.Context())
	if err != nil {
		slog.Error("list cameras for recording days", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	cameraIDs := make([]string, 0, len(cameras))
	for _, camera := range cameras {
		if camera.Enabled {
			cameraIDs = append(cameraIDs, camera.ID)
		}
	}

	days, err := s.Recording.ListDays(r.Context(), cameraIDs, params.From, params.To, location)
	if err != nil {
		slog.Error("list recording days", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	response, err := mapRecordingDays(days, location)
	if err != nil {
		slog.Error("map recording days", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	writeJSON(w, http.StatusOK, response)
}

// PostCameraPlayback mints a short-lived playback session for recorded video.
// Any authenticated user; 404 if the camera does not exist.
func (s *Server) PostCameraPlayback(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if auth.UserFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "unauthorized")
		return
	}
	if s.Media == nil || s.Recording == nil {
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

	segment, err := s.Recording.FindAt(r.Context(), id.String(), body.Start)
	if err != nil {
		if errors.Is(err, recording.ErrNotFound) || errors.Is(err, recording.ErrOutsideRecording) {
			writeError(w, http.StatusNotFound, "recording not found at requested time", "not_found")
			return
		}
		slog.Error("find recording for playback", "err", err, "camera_id", id.String())
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	_, localAvailable, err := s.Recording.LocalPath(segment)
	if err != nil {
		slog.Error("check local recording for playback", "err", err, "recording_id", segment.ID)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	var (
		session            media.PlaybackStream
		source             PlaybackSessionSource
		offset             float32
		nextRecordingStart *time.Time
	)
	if localAvailable {
		session, err = s.Media.IssuePlayback(r.Context(), id.String(), body.Start, durationSec)
		source = Local
	} else if driveFileID(segment.ArchiveLocation) != "" && s.DrivePlayback != nil {
		next, nextErr := s.Recording.NextAfter(r.Context(), id.String(), segment.StartedAt)
		if nextErr == nil {
			nextRecordingStart = &next.StartedAt
		} else if !errors.Is(nextErr, recording.ErrNotFound) {
			slog.Error("find next recording for playback", "err", nextErr, "recording_id", segment.ID)
			writeError(w, http.StatusInternalServerError, "internal error", "internal")
			return
		}
		session, err = s.Media.IssueArchivedPlayback(r.Context(), id.String(), segment.ID)
		source = Gdrive
		offset = float32(max(0, body.Start.Sub(segment.StartedAt).Seconds()))
	} else {
		writeError(w, http.StatusConflict, "recording is no longer available locally or in Drive", "unavailable")
		return
	}
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
		CameraId:           session.CameraID,
		RecordingId:        segment.ID,
		PlaybackUrl:        session.PlaybackURL,
		Token:              session.Token,
		ExpiresAt:          session.ExpiresAt,
		Source:             source,
		StartOffsetSec:     offset,
		NextRecordingStart: nextRecordingStart,
	})
}

// GetRecordingContent proxies a Drive-archived MP4 into the native timeline
// player. The short-lived query token is independent of the UI's cookie/Bearer
// session so remote dashboards and native video range requests both work.
func (s *Server) GetRecordingContent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params GetRecordingContentParams) {
	recordingID := id.String()
	if s.Media == nil || !s.Media.ValidateArchivedPlayback(params.Token, recordingID) {
		writeError(w, http.StatusUnauthorized, "invalid or expired playback token", "unauthorized")
		return
	}
	if s.Recording == nil || s.DrivePlayback == nil {
		writeError(w, http.StatusNotFound, "Drive archive is unavailable", "not_found")
		return
	}
	segment, err := s.Recording.Get(r.Context(), recordingID)
	if err != nil {
		if errors.Is(err, recording.ErrNotFound) {
			writeError(w, http.StatusNotFound, "recording not found", "not_found")
			return
		}
		slog.Error("get archived recording", "err", err, "recording_id", recordingID)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	fileID := driveFileID(segment.ArchiveLocation)
	if fileID == "" {
		writeError(w, http.StatusNotFound, "recording has no Drive archive", "not_found")
		return
	}
	byteRange := strings.TrimSpace(r.Header.Get("Range"))
	if !validByteRange(byteRange) {
		writeError(w, http.StatusRequestedRangeNotSatisfiable, "invalid byte range", "invalid_range")
		return
	}

	upstream, err := s.DrivePlayback.Download(r.Context(), fileID, byteRange)
	if err != nil {
		switch {
		case errors.Is(err, gdrive.ErrArchiveNotFound):
			writeError(w, http.StatusNotFound, "Drive archive not found", "not_found")
		case errors.Is(err, gdrive.ErrRangeNotSatisfiable):
			writeError(w, http.StatusRequestedRangeNotSatisfiable, "byte range not satisfiable", "invalid_range")
		default:
			slog.Error("stream Drive recording", "err", err, "recording_id", recordingID)
			writeError(w, http.StatusBadGateway, "Drive archive could not be read", "archive_unavailable")
		}
		return
	}
	defer upstream.Body.Close()

	for _, header := range []string{"Content-Length", "Content-Range", "ETag", "Last-Modified"} {
		if value := upstream.Header.Get(header); value != "" {
			w.Header().Set(header, value)
		}
	}
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Disposition", "inline")
	status := upstream.StatusCode
	if status != http.StatusOK && status != http.StatusPartialContent {
		status = http.StatusBadGateway
	}
	w.WriteHeader(status)
	if _, err := io.Copy(w, upstream.Body); err != nil && r.Context().Err() == nil {
		slog.Debug("Drive recording stream ended", "err", err, "recording_id", recordingID)
	}
}

func driveFileID(location *string) string {
	if location == nil {
		return ""
	}
	fileID, ok := strings.CutPrefix(strings.TrimSpace(*location), "gdrive:")
	if !ok {
		return ""
	}
	return strings.TrimSpace(fileID)
}

func validByteRange(value string) bool {
	if value == "" {
		return true
	}
	if len(value) > 128 || !strings.HasPrefix(value, "bytes=") || strings.Contains(value, ",") {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(value, "bytes="), "-")
	if len(parts) != 2 || parts[0] == "" && parts[1] == "" {
		return false
	}
	for _, part := range parts {
		if part == "" {
			continue
		}
		if _, err := strconv.ParseUint(part, 10, 64); err != nil {
			return false
		}
	}
	return true
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

func mapRecordingDays(days []recording.DayAvailability, location *time.Location) (RecordingDayList, error) {
	out := RecordingDayList{Days: make([]RecordingDayAvailability, 0, len(days))}
	for _, day := range days {
		date, err := time.ParseInLocation(time.DateOnly, day.Date, location)
		if err != nil {
			return RecordingDayList{}, fmt.Errorf("parse recording day %q: %w", day.Date, err)
		}
		out.Days = append(out.Days, RecordingDayAvailability{
			Date:   openapi_types.Date{Time: date},
			Source: RecordingDaySource(day.Source),
		})
	}
	return out, nil
}
