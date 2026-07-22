package camera

import "time"

// Status values matching the cameras.status CHECK constraint.
const (
	StatusOnline  = "online"
	StatusOffline = "offline"
	StatusUnknown = "unknown"
)

// Stream role values matching stream_profiles.role.
const (
	RoleLive   = "live"
	RoleRecord = "record"
)

// DefaultDriver is the phase-2 generic RTSP driver identifier.
const DefaultDriver = "generic-rtsp"

// Camera is the domain camera model (no secrets).
type Camera struct {
	ID             string
	Name           string
	Driver         string
	Host           string
	Username       string
	Enabled        bool
	Status         string
	StreamProfiles []StreamProfile
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// StreamProfile is a live or record RTSP profile.
type StreamProfile struct {
	ID      string
	Role    string
	RTSPURL string
	Codec   *string
	Width   *int
	Height  *int
}

// CreateInput is the payload for creating a camera.
type CreateInput struct {
	Name          string
	Driver        string
	Host          string
	Username      string
	Password      string
	Enabled       bool
	LiveRTSPURL   string
	RecordRTSPURL string
}

// UpdateInput is a partial update; nil pointer means leave unchanged.
type UpdateInput struct {
	Name          *string
	Driver        *string
	Host          *string
	Username      *string
	Password      *string
	Enabled       *bool
	LiveRTSPURL   *string
	RecordRTSPURL *string
}
