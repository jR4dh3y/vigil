package event

import "time"

// Severity values matching events.severity CHECK and OpenAPI Event.severity.
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)

// Common event types produced by the system.
const (
	TypeCameraOnline    = "camera.online"
	TypeCameraOffline   = "camera.offline"
	TypeDiskLow         = "disk.low"
	TypeArchiveComplete = "archive.complete"
)

// Event is the domain event model.
type Event struct {
	ID            string
	CameraID      *string
	Type          string
	Severity      string
	Title         string
	Message       string
	StartedAt     time.Time
	EndedAt       *time.Time
	Acknowledged  bool
	Metadata      map[string]any
	ThumbnailPath *string
	CreatedAt     time.Time
}

// EventInput is the payload for emitting a new event.
type EventInput struct {
	CameraID  *string
	Type      string
	Severity  string
	Title     string
	Message   string
	StartedAt time.Time
	EndedAt   *time.Time
	Metadata  map[string]any
}

// Cursor identifies one event's position in the newest-first list order.
type Cursor struct {
	StartedAt time.Time
	ID        string
}

// ListFilter controls ListEvents query options.
type ListFilter struct {
	Limit              int
	Before             *time.Time
	Cursor             *Cursor
	CameraID           *string
	Type               *string
	UnacknowledgedOnly bool
}
