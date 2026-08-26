package jobs

import (
	"testing"
	"time"

	"github.com/nvr/nvr/server/internal/config"
)

func TestNudgeArchiveCoalescesBeforeStart(t *testing.T) {
	s := NewScheduler(Config{})
	for range 3 {
		s.NudgeArchive()
	}
	if len(s.nudgeCh) != 1 {
		t.Fatalf("nudge channel holds %d signals, want coalesced 1", len(s.nudgeCh))
	}
}

func TestStopWithoutStartIsSafe(t *testing.T) {
	s := NewScheduler(Config{})
	s.Stop() // must not panic or block without a running loop
}

func TestArchiveIntervalDefaultsAndClamp(t *testing.T) {
	def := NewScheduler(Config{})
	if def.cfg.ArchiveInterval != config.DefaultArchiveInterval {
		t.Fatalf("default interval = %v, want %v", def.cfg.ArchiveInterval, config.DefaultArchiveInterval)
	}
	tight := NewScheduler(Config{ArchiveInterval: time.Second})
	if tight.cfg.ArchiveInterval != config.MinArchiveInterval {
		t.Fatalf("clamped interval = %v, want %v", tight.cfg.ArchiveInterval, config.MinArchiveInterval)
	}
}
