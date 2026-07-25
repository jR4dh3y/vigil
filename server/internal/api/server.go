package api

import (
	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/camera"
	"github.com/nvr/nvr/server/internal/event"
	"github.com/nvr/nvr/server/internal/media"
	"github.com/nvr/nvr/server/internal/recording"
	"github.com/nvr/nvr/server/internal/storage/gdrive"
	"github.com/nvr/nvr/server/internal/store"
)

// Server implements ServerInterface.
type Server struct {
	Queries       *store.Queries
	Auth          *auth.Service
	Camera        *camera.Service
	Media         *media.Service
	Recording     *recording.Service
	Event         *event.Service
	GDrive        *gdrive.Service
	Version       string
	Commit        string
	RecordingsDir string
	// DefaultRetentionDays is used when settings has no retentionDays key.
	DefaultRetentionDays int
}

// NewServer constructs an API server.
func NewServer(
	q *store.Queries,
	authSvc *auth.Service,
	camSvc *camera.Service,
	mediaSvc *media.Service,
	recordingSvc *recording.Service,
	eventSvc *event.Service,
	gdriveSvc *gdrive.Service,
	version, commit string,
	recordingsDir string,
	defaultRetentionDays int,
) *Server {
	if defaultRetentionDays <= 0 {
		defaultRetentionDays = recording.DefaultRetentionDays
	}
	return &Server{
		Queries:              q,
		Auth:                 authSvc,
		Camera:               camSvc,
		Media:                mediaSvc,
		Recording:            recordingSvc,
		Event:                eventSvc,
		GDrive:               gdriveSvc,
		Version:              version,
		Commit:               commit,
		RecordingsDir:        recordingsDir,
		DefaultRetentionDays: defaultRetentionDays,
	}
}
