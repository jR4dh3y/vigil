package storage

import (
	"context"
	"io"
	"time"
)

// Provider is the archive/object storage seam (arch.md §13).
// Implementations: s3-compatible, local, and later Google Drive.
type Provider interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]ObjectInfo, error)
	PresignURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

// ObjectInfo describes a stored object returned by List.
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
}
