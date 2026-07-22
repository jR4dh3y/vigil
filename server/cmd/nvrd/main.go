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
	mediaSvc := media.NewService(media.Config{
		APIURL:        cfg.MediaMTXAPIURL,
		WebRTCURL:     cfg.MediaMTXWEBRTCURL,
		HLSURL:        cfg.MediaMTXHLSURL,
		PlaybackURL:   cfg.MediaMTXPlaybackURL,
		RecordingsDir: cfg.RecordingsDir,
	}, cameraSvc)

	retentionDays := cfg.RetentionDays
	if days, ok := loadRetentionFromDB(queries); ok {
		retentionDays = days
	}
	recordingSvc := recording.NewService(queries, recording.Config{
		RecordingsDir: cfg.RecordingsDir,
		RetentionDays: retentionDays,
	})

	eventBus := event.NewBus()
	eventSvc := event.NewService(queries, eventBus)

	scheduler := jobs.NewScheduler(jobs.Config{
		Cameras:       cameraSvc,
		Events:        eventSvc,
		Recording:     recordingSvc,
		RecordingsDir: cfg.RecordingsDir,
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
		cfg.Version,
		cfg.Commit,
		cfg.RecordingsDir,
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
			"recordings", cfg.RecordingsDir,
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
	row, err := q.GetSetting(context.Background(), "retentionDays")
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			slog.Warn("load retention setting", "err", err)
		}
		return 0, false
	}
	n, err := strconv.Atoi(row.Value)
	if err != nil || n < 1 {
		return 0, false
	}
	return n, true
}
