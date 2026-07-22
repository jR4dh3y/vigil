package config

import (
	"log/slog"
	"os"
)

type Config struct {
	HTTPAddr string
	DataDir  string
	LogLevel slog.Level
}

func Load() (*Config, error) {
	return &Config{
		HTTPAddr: env("NVR_HTTP_ADDR", ":8080"),
		DataDir:  env("NVR_DATA_DIR", "./data"),
		LogLevel: slog.LevelInfo,
	}, nil
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
