package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nvr/nvr/server/internal/camera"
	"github.com/nvr/nvr/server/internal/event"
	"github.com/nvr/nvr/server/internal/recording"
	"github.com/nvr/nvr/server/internal/storage/gdrive"
	"github.com/robfig/cron/v3"
	"github.com/shirou/gopsutil/v4/disk"
)

// Config holds scheduler dependencies and options.
type Config struct {
	Cameras       *camera.Service
	Events        *event.Service
	Recording     *recording.Service
	GDrive        *gdrive.Service
	RecordingsDir string
	// HealthTimeout bounds each camera health probe (default 8s).
	HealthTimeout time.Duration
	// DiskThreshold is the used-percent at which disk.low is emitted (default 90).
	DiskThreshold float64
}

// Scheduler runs periodic background tasks via robfig/cron.
type Scheduler struct {
	cfg           Config
	cron          *cron.Cron
	mu            sync.Mutex
	lastDiskAlert time.Time
}

// NewScheduler constructs a jobs Scheduler. Call Start to begin cron jobs.
func NewScheduler(cfg Config) *Scheduler {
	if cfg.HealthTimeout <= 0 {
		cfg.HealthTimeout = 8 * time.Second
	}
	if cfg.DiskThreshold <= 0 {
		cfg.DiskThreshold = 90
	}
	return &Scheduler{
		cfg: cfg,
		// All schedules (including midnight archive) use UTC for predictable ops.
		cron: cron.New(cron.WithSeconds(), cron.WithLocation(time.UTC)),
	}
}

// Start registers cron jobs and starts the scheduler. It is safe to call once.
func (s *Scheduler) Start() error {
	// Every minute: camera health probe.
	if _, err := s.cron.AddFunc("0 * * * * *", s.safeRun("camera_health", s.probeCameras)); err != nil {
		return err
	}
	// Every hour: retention prune.
	if _, err := s.cron.AddFunc("0 0 * * * *", s.safeRun("retention_prune", s.pruneRecordings)); err != nil {
		return err
	}
	// Every 5 minutes: disk usage check.
	if _, err := s.cron.AddFunc("0 */5 * * * *", s.safeRun("disk_check", s.checkDisk)); err != nil {
		return err
	}
	// Midnight UTC: archive unarchived recordings to Google Drive (long timeout).
	if _, err := s.cron.AddFunc("0 0 0 * * *", s.safeRunTimeout("gdrive_archive", 30*time.Minute, s.archiveToGDrive)); err != nil {
		return err
	}
	s.cron.Start()
	slog.Info("jobs scheduler started", "cron_location", "UTC")
	return nil
}

// Stop gracefully stops cron jobs, waiting for in-flight work (incl. long archive runs).
func (s *Scheduler) Stop() {
	if s.cron == nil {
		return
	}
	ctx := s.cron.Stop()
	// Archive job may run up to 30m; wait a bit longer than default short jobs.
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Minute):
		slog.Warn("jobs scheduler stop timed out")
	}
	slog.Info("jobs scheduler stopped")
}

func (s *Scheduler) safeRun(name string, fn func(context.Context)) func() {
	return s.safeRunTimeout(name, 2*time.Minute, fn)
}

func (s *Scheduler) safeRunTimeout(name string, timeout time.Duration, fn func(context.Context)) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("job panicked", "job", name, "recover", r)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		fn(ctx)
	}
}

func (s *Scheduler) probeCameras(ctx context.Context) {
	if s.cfg.Cameras == nil || s.cfg.Events == nil {
		return
	}
	list, err := s.cfg.Cameras.List(ctx)
	if err != nil {
		slog.Warn("camera health: list failed", "err", err)
		return
	}
	for _, cam := range list {
		if ctx.Err() != nil {
			return
		}
		if !cam.Enabled {
			// Disabled cameras are not probed; leave status as-is.
			continue
		}
		s.probeOne(ctx, cam)
	}
}

func (s *Scheduler) probeOne(ctx context.Context, cam camera.Camera) {
	probeCtx, cancel := context.WithTimeout(ctx, s.cfg.HealthTimeout)
	defer cancel()

	source, err := s.cfg.Cameras.LiveSourceURL(probeCtx, cam.ID)
	online := false
	if err == nil && source != "" {
		result, probeErr := s.cfg.Cameras.Probe(probeCtx, source, "", "")
		if probeErr == nil && result.Reachable {
			online = true
		}
	}

	newStatus := camera.StatusOffline
	if online {
		newStatus = camera.StatusOnline
	}
	if cam.Status == newStatus {
		return
	}

	if err := s.cfg.Cameras.SetStatus(ctx, cam.ID, newStatus); err != nil {
		slog.Warn("camera health: set status failed", "camera_id", cam.ID, "err", err)
		return
	}

	evType := event.TypeCameraOffline
	severity := event.SeverityWarning
	title := "Camera offline"
	message := cam.Name + " is offline"
	if online {
		evType = event.TypeCameraOnline
		severity = event.SeverityInfo
		title = "Camera online"
		message = cam.Name + " is online"
	}
	camID := cam.ID
	if _, err := s.cfg.Events.Emit(ctx, event.EventInput{
		CameraID:  &camID,
		Type:      evType,
		Severity:  severity,
		Title:     title,
		Message:   message,
		StartedAt: time.Now().UTC(),
		Metadata: map[string]any{
			"previousStatus": cam.Status,
			"status":         newStatus,
		},
	}); err != nil {
		slog.Warn("camera health: emit event failed", "camera_id", cam.ID, "err", err)
	}
}

func (s *Scheduler) pruneRecordings(ctx context.Context) {
	if s.cfg.Recording == nil {
		return
	}
	var (
		n   int64
		err error
	)
	if s.cfg.GDrive != nil && s.cfg.GDrive.HasStoredConnection(ctx) {
		// A linked archive target makes archive-before-prune the durable policy:
		// failed/pending uploads remain indexed for the next retry.
		n, err = s.cfg.Recording.PruneArchived(ctx)
	} else {
		n, err = s.cfg.Recording.Prune(ctx)
	}
	if err != nil {
		slog.Warn("retention prune failed", "err", err)
		return
	}
	if n > 0 {
		slog.Info("retention prune complete", "deleted", n)
	}
}

func (s *Scheduler) checkDisk(ctx context.Context) {
	if s.cfg.Events == nil {
		return
	}
	path := s.cfg.RecordingsDir
	if s.cfg.Recording != nil {
		if dir := s.cfg.Recording.RecordingsDir(); dir != "" {
			path = dir
		}
	}
	if path == "" {
		return
	}
	usage, err := disk.UsageWithContext(ctx, path)
	if err != nil {
		slog.Warn("disk check failed", "path", path, "err", err)
		return
	}
	if usage.UsedPercent < s.cfg.DiskThreshold {
		return
	}

	s.mu.Lock()
	if time.Since(s.lastDiskAlert) < time.Hour {
		s.mu.Unlock()
		return
	}
	s.lastDiskAlert = time.Now()
	s.mu.Unlock()

	if _, err := s.cfg.Events.Emit(ctx, event.EventInput{
		Type:      event.TypeDiskLow,
		Severity:  event.SeverityCritical,
		Title:     "Disk space low",
		Message:   fmt.Sprintf("Recordings volume is above %.0f%% used", s.cfg.DiskThreshold),
		StartedAt: time.Now().UTC(),
		Metadata: map[string]any{
			"path":        path,
			"usedPercent": usage.UsedPercent,
			"totalBytes":  usage.Total,
			"freeBytes":   usage.Free,
		},
	}); err != nil {
		slog.Warn("disk check: emit event failed", "err", err)
	}
}
