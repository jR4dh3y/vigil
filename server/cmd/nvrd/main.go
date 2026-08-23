package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	"github.com/nvr/nvr/server/internal/bootstrap"
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
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "setup":
			// The setup command never starts the server and never blocks on a TTY.
			os.Exit(runSetup(os.Args[2:]))
		case "serve", "server":
			// Explicit server command; fall through to normal startup.
		default:
			fmt.Fprintf(os.Stderr, "nvrd: unknown command %q\n", os.Args[1])
			os.Exit(2)
		}
	}

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

	// First-run admin bootstrap from paired env vars (inert once a user exists).
	if err := bootstrapEnvAdmin(ctx, queries); err != nil {
		slog.Error("admin bootstrap", "err", err)
		os.Exit(1)
	}

	// Resolve env-over-DB precedence for public and hosted dashboard URLs.
	dbPublicURL, _ := loadSettingFromDB(queries, bootstrap.SettingPublicURL)
	dbHostedURL, _ := loadSettingFromDB(queries, bootstrap.SettingHostedDashboardURL)
	publicURL := bootstrap.ResolveURL(cfg.PublicURL, dbPublicURL)
	hostedURL := bootstrap.ResolveURL(cfg.HostedDashboardURL, dbHostedURL)

	// Validate configured URLs; empty means "not configured".
	for name, val := range map[string]string{
		"NVR_PUBLIC_URL":           publicURL,
		"NVR_HOSTED_DASHBOARD_URL": hostedURL,
	} {
		if val == "" {
			continue
		}
		if err := bootstrap.ValidateURL(val); err != nil {
			slog.Error("invalid URL config", "var", name, "err", err)
			os.Exit(1)
		}
	}

	corsOrigins, err := bootstrap.CORSOrigins(cfg.CORSOrigins, hostedURL)
	if err != nil {
		slog.Error("CORS origins", "err", err)
		os.Exit(1)
	}

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
	retentionDays := cfg.RetentionDays
	if days, ok := loadRetentionFromDB(queries); ok {
		retentionDays = days
	}

	mediaSvc := media.NewService(media.Config{
		APIURL:           cfg.MediaMTXAPIURL,
		WebRTCURL:        cfg.MediaMTXWEBRTCURL,
		HLSURL:           cfg.MediaMTXHLSURL,
		PlaybackURL:      cfg.MediaMTXPlaybackURL,
		RecordingsDir:    recordingsDir,
		RecordingEnabled: recordingEnabled,
		RetentionDays:    retentionDays,
	}, cameraSvc)

	// MediaMTX path configuration is runtime state and is lost whenever its
	// container restarts. Restore all enabled cameras during NVR startup so
	// continuous recording does not depend on someone opening the dashboard.
	cameras, err := cameraSvc.List(ctx)
	if err != nil {
		slog.Warn("list cameras for mediamtx startup sync", "err", err)
	} else {
		if err := mediaSvc.ReapplyCameraPaths(ctx, cameras); err != nil {
			slog.Warn("mediamtx startup sync failed", "err", err)
		}
	}

	recordingSvc := recording.NewService(queries, recording.Config{
		RecordingsDir: recordingsDir,
		RetentionDays: retentionDays,
	})
	go func() {
		stats, err := recordingSvc.ReconcileDisk(ctx, 5*time.Second)
		if err != nil {
			slog.Warn("recording startup reconciliation failed", "err", err)
			return
		}
		slog.Info("recording startup reconciliation complete",
			"discovered", stats.Discovered,
			"indexed", stats.Indexed,
			"existing", stats.Existing,
			"skipped", stats.Skipped,
			"failed", stats.Failed,
		)
		cleanupCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		cleanup, err := recordingSvc.CleanupArchivedLocals(cleanupCtx, 5*time.Second)
		if err != nil {
			slog.Warn("archived local startup cleanup failed", "err", err)
			return
		}
		if cleanup.Matched > 0 || cleanup.Failed > 0 {
			slog.Info("archived local startup cleanup complete",
				"scanned", cleanup.Scanned,
				"matched", cleanup.Matched,
				"deleted", cleanup.Deleted,
				"failed", cleanup.Failed,
			)
		}
	}()

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
		retentionDays,
	)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(api.CORSMiddleware(corsOrigins))
	r.Use(api.SessionMiddleware(authSvc))

	// MediaMTX external auth hook (not public OpenAPI; bind via authHTTPAddress).
	r.Post("/internal/mediamtx/auth", mediaSvc.AuthHandler())
	// MediaMTX runOnRecordSegmentComplete hook → recording index.
	r.Post("/internal/mediamtx/segment-complete", recordingSvc.SegmentCompleteHandler())

	// WebSocket event stream (not in OpenAPI).
	r.Get("/api/v1/ws", apiServer.HandleWS)

	// OpenAPI routes under /api/v1
	api.RegisterOpenAPIRoutes(apiServer, r, "/api/v1")

	// Static SPA (full) or connection page (slim) for everything else
	r.NotFound(ui.Handler(ui.Config{
		PublicURL:          publicURL,
		HostedDashboardURL: hostedURL,
	}).ServeHTTP)

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

// bootstrapEnvAdmin applies paired NVR_ADMIN_USERNAME/NVR_ADMIN_PASSWORD only
// when the database has no users. Partial or invalid configuration is fatal only
// while setup is required; once any user exists the env values are ignored.
func bootstrapEnvAdmin(ctx context.Context, q *store.Queries) error {
	username, hasUsername := os.LookupEnv("NVR_ADMIN_USERNAME")
	password, hasPassword := os.LookupEnv("NVR_ADMIN_PASSWORD")
	if !hasUsername && !hasPassword {
		return nil
	}

	count, err := q.CountUsers(ctx)
	if err != nil {
		return fmt.Errorf("count users: %w", err)
	}
	if count > 0 {
		// Setup already complete; env bootstrap is inert.
		return nil
	}

	// Setup is required: partial or invalid configuration is a startup error.
	if username == "" || password == "" {
		return errors.New("NVR_ADMIN_USERNAME and NVR_ADMIN_PASSWORD must both be set for first-run bootstrap")
	}
	if _, err := auth.CreateFirstAdmin(ctx, q, username, password); err != nil {
		return fmt.Errorf("first-run bootstrap: %w", err)
	}
	slog.Info("created first admin from environment")
	return nil
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
