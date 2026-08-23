package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/event"
	"github.com/nvr/nvr/server/internal/store"
)

func TestGeneratedParameterErrorsUseJSONContract(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "invalid event ID", url: "/api/v1/events/not-a-uuid"},
		{
			name: "invalid recording camera ID",
			url:  "/api/v1/recordings/days?from=2026-08-01T00%3A00%3A00Z&to=2026-08-02T00%3A00%3A00Z&timeZone=UTC&cameraId=not-a-uuid",
		},
		{name: "missing required recording range", url: "/api/v1/recordings/days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			RegisterOpenAPIRoutes(&Server{}, router, "/api/v1")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.url, nil))

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
				t.Fatalf("content-type=%q", contentType)
			}
			var response Error
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Code == nil || *response.Code != "validation" || response.Error == "" {
				t.Fatalf("unexpected response: %+v", response)
			}
		})
	}
}

func TestAuthStatusAllowsAnonymousRequests(t *testing.T) {
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatal(err)
	}

	router := chi.NewRouter()
	RegisterOpenAPIRoutes(&Server{Queries: store.New(db)}, router, "/api/v1")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/auth/status", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response AuthStatus
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !response.SetupRequired || response.User != nil {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestProtectedEventRouteAcceptsBearerSession(t *testing.T) {
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatal(err)
	}

	queries := store.New(db)
	user, err := auth.CreateFirstAdmin(context.Background(), queries, "admin", "password123")
	if err != nil {
		t.Fatal(err)
	}
	authService := auth.NewService(queries)
	token, _, err := authService.CreateSession(context.Background(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	eventService := event.NewService(queries, nil)
	emitted, err := eventService.Emit(context.Background(), event.EventInput{
		Type:      event.TypeDiskLow,
		Severity:  event.SeverityWarning,
		Title:     "Disk space is low",
		StartedAt: time.Date(2026, 8, 23, 8, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	router := chi.NewRouter()
	router.Use(SessionMiddleware(authService))
	RegisterOpenAPIRoutes(&Server{Queries: queries, Auth: authService, Event: eventService}, router, "/api/v1")
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/events/"+emitted.ID, nil)
	request.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response Event
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Id != emitted.ID {
		t.Fatalf("event ID=%q", response.Id)
	}
}
