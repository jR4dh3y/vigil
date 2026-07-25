package api

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/nvr/nvr/server/internal/storage"
	"github.com/nvr/nvr/server/internal/storage/gdrive"
)

// GetGDriveStatus returns Google Drive OAuth configuration and connection status.
func (s *Server) GetGDriveStatus(w http.ResponseWriter, r *http.Request) {
	if requireUser(w, r) == nil {
		return
	}
	if s.GDrive == nil {
		writeJSON(w, http.StatusOK, GDriveStatus{
			Configured: false,
			Connected:  false,
		})
		return
	}
	st, err := s.GDrive.Status(r.Context())
	if err != nil {
		slog.Error("gdrive status", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	out := GDriveStatus{
		Configured: st.Configured,
		Connected:  st.Connected,
	}
	if st.ConnectionError != "" {
		connectionError := st.ConnectionError
		out.ConnectionError = &connectionError
	}
	if st.AccountEmail != "" {
		email := st.AccountEmail
		out.AccountEmail = &email
	}
	if st.ConnectedAt != "" {
		if t, err := time.Parse(time.RFC3339, st.ConnectedAt); err == nil {
			out.ConnectedAt = &t
		}
	}
	if st.FolderID != "" {
		folder := st.FolderID
		out.FolderId = &folder
	}
	writeJSON(w, http.StatusOK, out)
}

// PostGDriveConnect starts the Google Drive OAuth flow (admin).
func (s *Server) PostGDriveConnect(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}
	if s.GDrive == nil {
		writeError(w, http.StatusBadRequest, "google drive oauth is not configured", "not_configured")
		return
	}
	authURL, err := s.GDrive.BeginConnect(r.Context())
	if err != nil {
		slog.Error("gdrive begin connect", "err", err)
		writeError(w, http.StatusBadRequest, "google drive is not ready to connect", "not_configured")
		return
	}
	writeJSON(w, http.StatusOK, GDriveConnectResponse{AuthorizationUrl: authURL})
}

// GetGDriveCallback handles the OAuth redirect (public) and redirects to the UI.
func (s *Server) GetGDriveCallback(w http.ResponseWriter, r *http.Request, params GetGDriveCallbackParams) {
	if s.GDrive == nil {
		http.Redirect(w, r, gdrive.RedirectURL(gdrive.CallbackResult{
			OK:      false,
			Message: "not configured",
		}), http.StatusFound)
		return
	}

	code := ""
	if params.Code != nil {
		code = *params.Code
	}
	state := ""
	if params.State != nil {
		state = *params.State
	}
	errParam := ""
	if params.Error != nil {
		errParam = *params.Error
	}
	errDesc := ""
	if params.ErrorDescription != nil {
		errDesc = *params.ErrorDescription
	}

	result, err := s.GDrive.HandleCallback(r.Context(), code, state, errParam, errDesc)
	if err != nil {
		slog.Error("gdrive callback", "err", err)
		http.Redirect(w, r, gdrive.RedirectURL(gdrive.CallbackResult{
			OK:      false,
			Message: "internal error",
		}), http.StatusFound)
		return
	}
	http.Redirect(w, r, gdrive.RedirectURL(result), http.StatusFound)
}

// DeleteGDriveDisconnect clears Drive tokens (admin).
func (s *Server) DeleteGDriveDisconnect(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}
	if s.GDrive == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := s.GDrive.Disconnect(r.Context()); err != nil {
		slog.Error("gdrive disconnect", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PostGDriveArchive runs a batch archive of unarchived recordings to Google Drive (admin).
func (s *Server) PostGDriveArchive(w http.ResponseWriter, r *http.Request) {
	if requireAdmin(w, r) == nil {
		return
	}
	if s.GDrive == nil {
		writeError(w, http.StatusBadRequest, "google drive is not configured", "not_configured")
		return
	}
	if s.Recording == nil {
		writeError(w, http.StatusBadRequest, "recording service is not configured", "not_configured")
		return
	}
	if !s.GDrive.Connected(r.Context()) {
		writeError(w, http.StatusBadRequest, "google drive is not connected", "not_connected")
		return
	}

	limit := 50
	if r.Body != nil && r.Body != http.NoBody {
		var body GDriveArchiveRequest
		// Empty body is allowed (optional requestBody).
		err := decodeJSON(r, &body)
		if err != nil && err != io.EOF {
			writeError(w, http.StatusBadRequest, "invalid request body", "bad_request")
			return
		}
		if err == nil && body.Limit != nil {
			if *body.Limit < 1 || *body.Limit > 500 {
				writeError(w, http.StatusBadRequest, "limit must be between 1 and 500", "validation")
				return
			}
			limit = *body.Limit
		}
	}

	stats, err := s.GDrive.ArchivePending(r.Context(), storage.RecordingArchiveIndex{
		Recording: s.Recording,
	}, limit)
	if err != nil {
		slog.Error("gdrive archive", "err", err)
		if errors.Is(err, gdrive.ErrArchiveInProgress) {
			writeError(w, http.StatusConflict, err.Error(), "archive_in_progress")
			return
		}
		writeError(w, http.StatusBadRequest, "archive failed", "archive_failed")
		return
	}
	writeJSON(w, http.StatusOK, GDriveArchiveResponse{
		Uploaded: stats.Uploaded,
		Failed:   stats.Failed,
		Skipped:  stats.Skipped,
	})
}
