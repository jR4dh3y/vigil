package camera

import (
	"context"
	"testing"

	"github.com/nvr/nvr/server/internal/store"
)

func setupTestService(t *testing.T) *Service {
	t.Helper()
	dir := t.TempDir()
	db, err := store.Open(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewService(db, "")
}

func TestCreateListGetDeleteCamera(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	c, err := svc.Create(ctx, CreateInput{
		Name:          "Front door",
		Host:          "192.168.1.10",
		Username:      "admin",
		Password:      "secret",
		Enabled:       true,
		LiveRTSPURL:   "rtsp://192.168.1.10/live",
		RecordRTSPURL: "rtsp://192.168.1.10/main",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.ID == "" || c.Name != "Front door" {
		t.Fatalf("unexpected camera: %+v", c)
	}
	if len(c.StreamProfiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(c.StreamProfiles))
	}

	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 camera, got %d", len(list))
	}

	got, err := svc.Get(ctx, c.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Host != "192.168.1.10" {
		t.Fatalf("host: %q", got.Host)
	}

	updated, err := svc.Update(ctx, c.ID, UpdateInput{
		Name: strPtr("Porch"),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "Porch" {
		t.Fatalf("name: %q", updated.Name)
	}

	if err := svc.Delete(ctx, c.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := svc.Get(ctx, c.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateFromRTSPHost(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	c, err := svc.Create(ctx, CreateInput{
		Name:    "Cam",
		Host:    "rtsp://10.0.0.5/stream1",
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(c.StreamProfiles) != 2 {
		t.Fatalf("expected live+record from rtsp host, got %d", len(c.StreamProfiles))
	}
	for _, p := range c.StreamProfiles {
		if p.RTSPURL != "rtsp://10.0.0.5/stream1" {
			t.Fatalf("profile %s url: %q", p.Role, p.RTSPURL)
		}
	}
}

func TestEncryptStoredAsPlainWhenNoKey(t *testing.T) {
	enc, err := EncryptCredential("", "pw")
	if err != nil {
		t.Fatal(err)
	}
	if enc != "plain:pw" {
		t.Fatalf("got %q", enc)
	}
}

func strPtr(s string) *string { return &s }
