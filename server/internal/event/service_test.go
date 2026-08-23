package event

import (
	"context"
	"testing"
	"time"

	"github.com/nvr/nvr/server/internal/store"
)

func TestListCursorDoesNotSkipEventsWithEqualTimestamps(t *testing.T) {
	db, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatal(err)
	}

	service := NewService(store.New(db), nil)
	startedAt := time.Date(2026, 8, 23, 8, 30, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		if _, err := service.Emit(context.Background(), EventInput{
			Type:      TypeDiskLow,
			Severity:  SeverityWarning,
			Title:     "Disk space is low",
			StartedAt: startedAt,
		}); err != nil {
			t.Fatal(err)
		}
	}

	firstPage, err := service.List(context.Background(), ListFilter{Limit: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(firstPage) != 3 {
		t.Fatalf("first page length=%d", len(firstPage))
	}

	cursor := Cursor{StartedAt: firstPage[2].StartedAt, ID: firstPage[2].ID}
	secondPage, err := service.List(context.Background(), ListFilter{Limit: 3, Cursor: &cursor})
	if err != nil {
		t.Fatal(err)
	}
	if len(secondPage) != 2 {
		t.Fatalf("second page length=%d", len(secondPage))
	}

	seen := make(map[string]struct{}, 5)
	for _, event := range append(firstPage, secondPage...) {
		if _, exists := seen[event.ID]; exists {
			t.Fatalf("event %s appeared in both pages", event.ID)
		}
		seen[event.ID] = struct{}{}
	}
	if len(seen) != 5 {
		t.Fatalf("events across pages=%d", len(seen))
	}
}
