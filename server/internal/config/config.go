package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Package-level build metadata; overridden via ldflags in release builds.
var (
	Version = "0.1.0"
	Commit  = "dev"
)

// DefaultRetentionDays is the local fallback retention period used when no
// persisted or environment override is configured.
const DefaultRetentionDays = 7

type Config struct {
	HTTPAddr      string
	DataDir       string
	RecordingsDir string
	SecretsKey    string
	LogLevel      slog.Level
	Version       string
	Commit        string

	// MediaMTX control + browser-facing playback bases (phase 3 media plane).
	MediaMTXAPIURL      string // NVR_MEDIAMTX_API_URL — Control API
	MediaMTXWEBRTCURL   string // NVR_MEDIAMTX_WEBRTC_URL — WHEP base
	MediaMTXHLSURL      string // NVR_MEDIAMTX_HLS_URL — HLS base
	MediaMTXPlaybackURL string // NVR_MEDIAMTX_PLAYBACK_URL — optional playback server

	// RetentionDays is how long recording index rows are kept (default 7).
	RetentionDays int

	// Google Drive OAuth (optional archive tier).
	GoogleClientID     string // NVR_GOOGLE_CLIENT_ID
	GoogleClientSecret string // NVR_GOOGLE_CLIENT_SECRET
	GoogleRedirectURL  string // NVR_GOOGLE_REDIRECT_URL

	// PublicURL is the externally reachable URL of this server (NVR_PUBLIC_URL).
	// When non-empty it overrides the persisted publicUrl setting.
	PublicURL string
	// HostedDashboardURL is the hosted dashboard that manages this server
	// (NVR_HOSTED_DASHBOARD_URL). When non-empty it overrides the persisted one.
	HostedDashboardURL string
	// CORSOrigins are extra exact HTTPS origins allowed cross-origin
	// (NVR_CORS_ORIGINS, comma-separated).
	CORSOrigins []string
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:            env("NVR_HTTP_ADDR", ":8080"),
		DataDir:             env("NVR_DATA_DIR", "./data"),
		RecordingsDir:       env("NVR_RECORDINGS_DIR", "./recordings"),
		SecretsKey:          env("NVR_SECRETS_KEY", ""),
		LogLevel:            parseLogLevel(env("NVR_LOG_LEVEL", "info")),
		Version:             Version,
		Commit:              Commit,
		MediaMTXAPIURL:      env("NVR_MEDIAMTX_API_URL", "http://127.0.0.1:9997"),
		MediaMTXWEBRTCURL:   env("NVR_MEDIAMTX_WEBRTC_URL", "http://127.0.0.1:8889"),
		MediaMTXHLSURL:      env("NVR_MEDIAMTX_HLS_URL", "http://127.0.0.1:8888"),
		MediaMTXPlaybackURL: env("NVR_MEDIAMTX_PLAYBACK_URL", ""),
		RetentionDays:       envInt("NVR_RETENTION_DAYS", DefaultRetentionDays),
		GoogleClientID:      env("NVR_GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:  env("NVR_GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:   env("NVR_GOOGLE_REDIRECT_URL", ""),
		PublicURL:           env("NVR_PUBLIC_URL", ""),
		HostedDashboardURL:  env("NVR_HOSTED_DASHBOARD_URL", ""),
		CORSOrigins:         splitList(env("NVR_CORS_ORIGINS", "")),
	}
	return cfg, nil
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// EnsureDirs creates DataDir and RecordingsDir if they do not exist.
func (c *Config) EnsureDirs() error {
	for _, dir := range []string{c.DataDir, c.RecordingsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}
	return nil
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
