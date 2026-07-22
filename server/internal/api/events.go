package api

import (
	"errors"
	"log/slog"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/nvr/nvr/server/internal/event"
)

// ListEvents returns recent events for any authenticated user.
func (s *Server) ListEvents(w http.ResponseWriter, r *http.Request, params ListEventsParams) {
	if requireUser(w, r) == nil {
		return
	}
	if s.Event == nil {
		writeError(w, http.StatusInternalServerError, "event service unavailable", "internal")
		return
	}

	filter := event.ListFilter{}
	if params.Limit != nil {
		filter.Limit = *params.Limit
	}
	if params.Before != nil {
		filter.Before = params.Before
	}
	if params.CameraId != nil {
		id := params.CameraId.String()
		filter.CameraID = &id
	}
	if params.Type != nil {
		filter.Type = params.Type
	}
	if params.UnacknowledgedOnly != nil {
		filter.UnacknowledgedOnly = *params.UnacknowledgedOnly
	}

	list, err := s.Event.List(r.Context(), filter)
	if err != nil {
		slog.Error("list events", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}

	out := make([]Event, 0, len(list))
	for _, e := range list {
		out = append(out, mapEvent(e))
	}
	writeJSON(w, http.StatusOK, EventList{Events: out})
}

// AcknowledgeEvent marks an event as acknowledged.
func (s *Server) AcknowledgeEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if requireUser(w, r) == nil {
		return
	}
	if s.Event == nil {
		writeError(w, http.StatusInternalServerError, "event service unavailable", "internal")
		return
	}

	ev, err := s.Event.Acknowledge(r.Context(), id.String())
	if err != nil {
		if errors.Is(err, event.ErrNotFound) {
			writeError(w, http.StatusNotFound, "event not found", "not_found")
			return
		}
		slog.Error("acknowledge event", "err", err, "event_id", id.String())
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	writeJSON(w, http.StatusOK, mapEvent(ev))
}

func mapEvent(e event.Event) Event {
	sev := EventSeverity(e.Severity)
	if !sev.Valid() {
		sev = Info
	}
	out := Event{
		Id:           e.ID,
		Type:         e.Type,
		Severity:     sev,
		Title:        e.Title,
		Message:      e.Message,
		StartedAt:    e.StartedAt,
		Acknowledged: e.Acknowledged,
		CreatedAt:    e.CreatedAt,
		CameraId:     e.CameraID,
		EndedAt:      e.EndedAt,
	}
	if len(e.Metadata) > 0 {
		m := e.Metadata
		out.Metadata = &m
	}
	return out
}
