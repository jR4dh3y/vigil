package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/event"
	"github.com/nvr/nvr/server/internal/store"
)

func setupEventServer(t *testing.T) (*Server, event.Event) {
	t.Helper()
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatal(err)
	}

	service := event.NewService(store.New(db), nil)
	ev, err := service.Emit(context.Background(), event.EventInput{
		Type:      event.TypeDiskLow,
		Severity:  event.SeverityWarning,
		Title:     "Disk space is low",
		Message:   "Less than ten percent remains",
		StartedAt: time.Date(2026, 8, 23, 8, 30, 0, 0, time.UTC),
		Metadata:  map[string]any{"usedPercent": 91.0},
	})
	if err != nil {
		t.Fatal(err)
	}
	return &Server{Event: service}, ev
}

func TestGetEventReturnsEvent(t *testing.T) {
	server, emitted := setupEventServer(t)
	req := httptest.NewRequestWithContext(
		auth.WithUser(t.Context(), &auth.User{ID: "user", Role: auth.RoleViewer}),
		http.MethodGet,
		"/api/v1/events/"+emitted.ID,
		nil,
	)
	rr := httptest.NewRecorder()

	server.GetEvent(rr, req, uuid.MustParse(emitted.ID))

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var response Event
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Id != emitted.ID || response.Type != emitted.Type || response.Severity != Warning {
		t.Fatalf("unexpected event: %+v", response)
	}
}

func TestGetEventReturnsNotFound(t *testing.T) {
	server, _ := setupEventServer(t)
	id := uuid.New()
	req := httptest.NewRequestWithContext(
		auth.WithUser(t.Context(), &auth.User{ID: "user", Role: auth.RoleViewer}),
		http.MethodGet,
		"/api/v1/events/"+id.String(),
		nil,
	)
	rr := httptest.NewRecorder()

	server.GetEvent(rr, req, id)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestGetEventRequiresAuthentication(t *testing.T) {
	server, emitted := setupEventServer(t)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/events/"+emitted.ID, nil)
	rr := httptest.NewRecorder()

	server.GetEvent(rr, req, uuid.MustParse(emitted.ID))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestListEventsReturnsCompositeCursor(t *testing.T) {
	server, emitted := setupEventServer(t)
	req := httptest.NewRequestWithContext(
		auth.WithUser(t.Context(), &auth.User{ID: "user", Role: auth.RoleViewer}),
		http.MethodGet,
		"/api/v1/events",
		nil,
	)
	recorder := httptest.NewRecorder()

	server.ListEvents(recorder, req, ListEventsParams{})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response EventList
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.NextCursor == nil || response.NextCursor.Id.String() != emitted.ID || !response.NextCursor.StartedAt.Equal(emitted.StartedAt) {
		t.Fatalf("unexpected cursor: %+v", response.NextCursor)
	}
}

func TestListEventsRejectsConflictingPagination(t *testing.T) {
	server, _ := setupEventServer(t)
	before := time.Date(2026, 8, 23, 8, 30, 0, 0, time.UTC)
	req := httptest.NewRequestWithContext(
		auth.WithUser(t.Context(), &auth.User{ID: "user", Role: auth.RoleViewer}),
		http.MethodGet,
		"/api/v1/events",
		nil,
	)
	recorder := httptest.NewRecorder()

	server.ListEvents(recorder, req, ListEventsParams{
		Before: &before,
		Cursor: &EventCursor{StartedAt: before, Id: uuid.New()},
	})

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response Error
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Code == nil || *response.Code != "validation" {
		t.Fatalf("unexpected response: %+v", response)
	}
}
