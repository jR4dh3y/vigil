package media

import (
	"testing"
)

func TestRecordOptionsUseConfiguredRetention(t *testing.T) {
	svc := NewService(Config{
		RecordingsDir:    t.TempDir(),
		RecordingEnabled: true,
		RetentionDays:    1,
	}, nil)

	if got := svc.recordOptionsForCamera("camera").DeleteAfter; got != "1d" {
		t.Fatalf("initial DeleteAfter = %q, want 1d", got)
	}
	svc.SetRetentionDays(3)
	if got := svc.recordOptionsForCamera("camera").DeleteAfter; got != "3d" {
		t.Fatalf("updated DeleteAfter = %q, want 3d", got)
	}

	defaultSvc := NewService(Config{
		RecordingsDir:    t.TempDir(),
		RecordingEnabled: true,
	}, nil)
	if got := defaultSvc.recordOptionsForCamera("camera").DeleteAfter; got != "7d" {
		t.Fatalf("default DeleteAfter = %q, want 7d", got)
	}
}
