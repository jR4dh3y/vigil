package event

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/store"
)

// ErrNotFound is returned when an event does not exist.
var ErrNotFound = errors.New("event not found")

// Service persists events and publishes them on the in-process bus.
type Service struct {
	q   *store.Queries
	bus *Bus
}

// NewService constructs an event Service.
func NewService(q *store.Queries, bus *Bus) *Service {
	if bus == nil {
		bus = NewBus()
	}
	return &Service{q: q, bus: bus}
}

// Bus returns the underlying event bus (for WebSocket fan-out).
func (s *Service) Bus() *Bus {
	return s.bus
}

// Emit inserts an event into the database and publishes it on the bus.
func (s *Service) Emit(ctx context.Context, in EventInput) (Event, error) {
	typ := strings.TrimSpace(in.Type)
	if typ == "" {
		return Event{}, fmt.Errorf("event type is required")
	}

	severity := strings.TrimSpace(in.Severity)
	if severity == "" {
		severity = SeverityInfo
	}
	switch severity {
	case SeverityInfo, SeverityWarning, SeverityCritical:
	default:
		return Event{}, fmt.Errorf("invalid severity %q", severity)
	}

	started := in.StartedAt
	if started.IsZero() {
		started = time.Now().UTC()
	} else {
		started = started.UTC()
	}

	meta := in.Metadata
	if meta == nil {
		meta = map[string]any{}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return Event{}, fmt.Errorf("marshal metadata: %w", err)
	}

	params := store.InsertEventParams{
		ID:           uuid.NewString(),
		Type:         typ,
		Severity:     severity,
		Title:        in.Title,
		Message:      in.Message,
		StartedAt:    formatTime(started),
		Metadata:     string(metaJSON),
		Acknowledged: 0,
	}
	if in.CameraID != nil && strings.TrimSpace(*in.CameraID) != "" {
		params.CameraID = sql.NullString{String: strings.TrimSpace(*in.CameraID), Valid: true}
	}
	if in.EndedAt != nil && !in.EndedAt.IsZero() {
		params.EndedAt = sql.NullString{String: formatTime(in.EndedAt.UTC()), Valid: true}
	}

	row, err := s.q.InsertEvent(ctx, params)
	if err != nil {
		return Event{}, fmt.Errorf("insert event: %w", err)
	}

	ev, err := toDomain(row)
	if err != nil {
		return Event{}, err
	}
	s.bus.Publish(ev)
	return ev, nil
}

// Get returns a single event by id.
func (s *Service) Get(ctx context.Context, id string) (Event, error) {
	row, err := s.q.GetEvent(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Event{}, ErrNotFound
		}
		return Event{}, fmt.Errorf("get event: %w", err)
	}
	return toDomain(row)
}

// List returns events matching filter, newest first.
func (s *Service) List(ctx context.Context, filter ListFilter) ([]Event, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	params := store.ListEventsParams{
		LimitCount: int64(limit),
	}
	if filter.CameraID != nil && strings.TrimSpace(*filter.CameraID) != "" {
		params.CameraID = sql.NullString{String: strings.TrimSpace(*filter.CameraID), Valid: true}
	}
	if filter.Type != nil && strings.TrimSpace(*filter.Type) != "" {
		params.EventType = sql.NullString{String: strings.TrimSpace(*filter.Type), Valid: true}
	}
	if filter.Before != nil && !filter.Before.IsZero() {
		params.Before = sql.NullString{String: formatTime(filter.Before.UTC()), Valid: true}
	}
	if filter.UnacknowledgedOnly {
		params.UnackedOnly = sql.NullInt64{Int64: 1, Valid: true}
	}

	rows, err := s.q.ListEvents(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	out := make([]Event, 0, len(rows))
	for _, row := range rows {
		ev, err := toDomain(row)
		if err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, nil
}

// Acknowledge marks an event as acknowledged.
func (s *Service) Acknowledge(ctx context.Context, id string) (Event, error) {
	row, err := s.q.AcknowledgeEvent(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Event{}, ErrNotFound
		}
		return Event{}, fmt.Errorf("acknowledge event: %w", err)
	}
	return toDomain(row)
}

func toDomain(row store.Event) (Event, error) {
	started, err := parseTime(row.StartedAt)
	if err != nil {
		return Event{}, fmt.Errorf("parse started_at: %w", err)
	}
	created, err := parseTime(row.CreatedAt)
	if err != nil {
		// SQLite datetime('now') may differ from RFC3339; try auth parser.
		if t, e2 := auth.ParseSQLiteTime(row.CreatedAt); e2 == nil {
			created = t
		} else {
			created = started
		}
	}

	ev := Event{
		ID:           row.ID,
		Type:         row.Type,
		Severity:     row.Severity,
		Title:        row.Title,
		Message:      row.Message,
		StartedAt:    started,
		Acknowledged: row.Acknowledged != 0,
		CreatedAt:    created,
		Metadata:     map[string]any{},
	}
	if row.CameraID.Valid {
		id := row.CameraID.String
		ev.CameraID = &id
	}
	if row.EndedAt.Valid && row.EndedAt.String != "" {
		if t, e := parseTime(row.EndedAt.String); e == nil {
			ev.EndedAt = &t
		}
	}
	if row.ThumbnailPath.Valid {
		p := row.ThumbnailPath.String
		ev.ThumbnailPath = &p
	}
	if strings.TrimSpace(row.Metadata) != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(row.Metadata), &m); err == nil && m != nil {
			ev.Metadata = m
		}
	}
	return ev, nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func parseTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.UTC); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("parse time %q", s)
}
