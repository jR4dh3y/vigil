package config

import (
	"os"
	"testing"
)

func setEnv(t *testing.T, k, v string) {
	t.Helper()
	t.Setenv(k, v)
}

func TestLoadRejectsEvictThresholdAbove100(t *testing.T) {
	setEnv(t, "NVR_LOCAL_EVICT_THRESHOLD", "101")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for threshold > 100")
	}
}

func TestLoadRejectsNegativeEvictThreshold(t *testing.T) {
	setEnv(t, "NVR_LOCAL_EVICT_THRESHOLD", "-1")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for negative threshold")
	}
}

func TestLoadAcceptsFractionalEvictThreshold(t *testing.T) {
	setEnv(t, "NVR_LOCAL_EVICT_THRESHOLD", "85.5")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LocalEvictThreshold != 85.5 {
		t.Fatalf("threshold = %v, want 85.5", cfg.LocalEvictThreshold)
	}
}

func TestLoadDefaultsEvictThresholdDisabled(t *testing.T) {
	if err := os.Unsetenv("NVR_LOCAL_EVICT_THRESHOLD"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LocalEvictThreshold != 0 {
		t.Fatalf("threshold = %v, want 0", cfg.LocalEvictThreshold)
	}
}
