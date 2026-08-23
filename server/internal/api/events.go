package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
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
	if params.Before != nil && params.Cursor != nil {
		writeError(w, http.StatusBadRequest, "before and cursor cannot be used together", "validation")
		return
	}
	if params.Cursor != nil {
		if params.Cursor.StartedAt.IsZero() || params.Cursor.Id == uuid.Nil {
			writeError(w, http.StatusBadRequest, "cursor must include a valid startedAt and id", "validation")
			return
		}
		filter.Cursor = &event.Cursor{
			StartedAt: params.Cursor.StartedAt,
			ID:        params.Cursor.Id.String(),
		}
	}
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
	response := EventList{Events: out}
	if last := lastEventCursor(list); last != nil {
		response.NextCursor = last
	}
	writeJSON(w, http.StatusOK, response)
}

func lastEventCursor(events []event.Event) *EventCursor {
	if len(events) == 0 {
		return nil
	}
	last := events[len(events)-1]
	id, err := uuid.Parse(last.ID)
	if err != nil {
		return nil
	}
	return &EventCursor{StartedAt: last.StartedAt, Id: id}
}

// GetEvent returns one event for any authenticated user.
func (s *Server) GetEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	if requireUser(w, r) == nil {
		return
	}
	if s.Event == nil {
		writeError(w, http.StatusInternalServerError, "event service unavailable", "internal")
		return
	}

	ev, err := s.Event.Get(r.Context(), id.String())
	if err != nil {
		if errors.Is(err, event.ErrNotFound) {
			writeError(w, http.StatusNotFound, "event not found", "not_found")
			return
		}
		slog.Error("get event", "err", err, "event_id", id.String())
		writeError(w, http.StatusInternalServerError, "internal error", "internal")
		return
	}
	writeJSON(w, http.StatusOK, mapEvent(ev))
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
