package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// MediaStore is the storage boundary for media payloads (images, video, audio, documents).
type MediaStore interface {
	Put(ctx context.Context, key, mimeType string, data []byte) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

// ResolveMediaURL produces a client-loadable media URL.
// When s3ObjectURL is configured, it points directly to the public S3 / CDN endpoint.
// Otherwise, it points to the application's serving endpoint for the given surface.
func ResolveMediaURL(key string, s3ObjectURL string, forSurface string) string {
	key = strings.TrimPrefix(key, "/")
	if s3ObjectURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s3ObjectURL, "/"), key)
	}
	if forSurface == "dashboard" {
		return fmt.Sprintf("/dashboard/api/media/%s", key)
	}
	return fmt.Sprintf("/api/v1/media/%s", key)
}
