package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/nvr/nvr/server/internal/api"
	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/camera"
	"github.com/nvr/nvr/server/internal/config"
	"github.com/nvr/nvr/server/internal/event"
	"github.com/nvr/nvr/server/internal/jobs"
	"github.com/nvr/nvr/server/internal/media"
	"github.com/nvr/nvr/server/internal/recording"
	"github.com/nvr/nvr/server/internal/storage/gdrive"
	"github.com/nvr/nvr/server/internal/store"
	"github.com/nvr/nvr/server/internal/ui"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	})))

	if err := cfg.EnsureDirs(); err != nil {
		slog.Error("ensure dirs", "err", err)
		os.Exit(1)
	}

	db, err := store.Open(cfg.DataDir)
	if err != nil {
		slog.Error("open database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := store.Migrate(db); err != nil {
		slog.Error("migrate", "err", err)
		os.Exit(1)
	}

	queries := store.New(db)
	authSvc := auth.NewService(queries)
	cameraSvc := camera.NewService(db, cfg.SecretsKey)

	recordingsDir := cfg.RecordingsDir
	if dir, ok := loadSettingFromDB(queries, "recordingsDir"); ok && strings.TrimSpace(dir) != "" {
		recordingsDir = strings.TrimSpace(dir)
	}
	recordingEnabled := true
	if en, ok := loadSettingFromDB(queries, "recordingEnabled"); ok {
		recordingEnabled = parseEnvBool(en, true)
	}
	if strings.TrimSpace(recordingsDir) == "" {
		recordingEnabled = false
	} else if err := os.MkdirAll(recordingsDir, 0o755); err != nil {
		slog.Error("ensure recordings dir", "path", recordingsDir, "err", err)
		os.Exit(1)
	}

	mediaSvc := media.NewService(media.Config{
		APIURL:           cfg.MediaMTXAPIURL,
		WebRTCURL:        cfg.MediaMTXWEBRTCURL,
		HLSURL:           cfg.MediaMTXHLSURL,
		PlaybackURL:      cfg.MediaMTXPlaybackURL,
		RecordingsDir:    recordingsDir,
		RecordingEnabled: recordingEnabled,
	}, cameraSvc)

	retentionDays := cfg.RetentionDays
	if days, ok := loadRetentionFromDB(queries); ok {
		retentionDays = days
	}
	recordingSvc := recording.NewService(queries, recording.Config{
		RecordingsDir: recordingsDir,
		RetentionDays: retentionDays,
	})

	eventBus := event.NewBus()
	eventSvc := event.NewService(queries, eventBus)

	gdriveSvc := gdrive.NewService(gdrive.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
	}, queries, cfg.SecretsKey)

	scheduler := jobs.NewScheduler(jobs.Config{
		Cameras:       cameraSvc,
		Events:        eventSvc,
		Recording:     recordingSvc,
		GDrive:        gdriveSvc,
		RecordingsDir: recordingsDir,
	})
	if err := scheduler.Start(); err != nil {
		slog.Error("start jobs scheduler", "err", err)
		os.Exit(1)
	}

	apiServer := api.NewServer(
		queries,
		authSvc,
		cameraSvc,
		mediaSvc,
		recordingSvc,
		eventSvc,
		gdriveSvc,
		cfg.Version,
		cfg.Commit,
		recordingsDir,
		cfg.RetentionDays,
	)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(api.CORSMiddleware)
	r.Use(api.SessionMiddleware(authSvc))

	// MediaMTX external auth hook (not public OpenAPI; bind via authHTTPAddress).
	r.Post("/internal/mediamtx/auth", mediaSvc.AuthHandler())
	// MediaMTX runOnRecordSegmentComplete hook → recording index.
	r.Post("/internal/mediamtx/segment-complete", recordingSvc.SegmentCompleteHandler())

	// WebSocket event stream (not in OpenAPI).
	r.Get("/api/v1/ws", apiServer.HandleWS)

	// OpenAPI routes under /api/v1
	api.HandlerFromMuxWithBaseURL(apiServer, r, "/api/v1")

	// Static SPA for everything else
	r.NotFound(ui.Handler().ServeHTTP)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("nvrd listening",
			"http", cfg.HTTPAddr,
			"version", cfg.Version,
			"commit", cfg.Commit,
			"recordings", recordingsDir,
			"recording_enabled", mediaSvc.RecordingEnabled(),
			"retention_days", recordingSvc.RetentionDays(),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("nvrd shutting down")

	scheduler.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "err", err)
		os.Exit(1)
	}
}

func loadRetentionFromDB(q *store.Queries) (int, bool) {
	val, ok := loadSettingFromDB(q, "retentionDays")
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil || n < 1 {
		return 0, false
	}
	return n, true
}

func loadSettingFromDB(q *store.Queries, key string) (string, bool) {
	row, err := q.GetSetting(context.Background(), key)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			slog.Warn("load setting", "key", key, "err", err)
		}
		return "", false
	}
	return row.Value, true
}

func parseEnvBool(val string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
