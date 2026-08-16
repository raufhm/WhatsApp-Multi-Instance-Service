package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DiskStore stores media files on the local filesystem rooted at baseDir.
type DiskStore struct {
	baseDir string
}

// NewDiskStore initializes a DiskStore at baseDir, creating the directory if needed.
func NewDiskStore(baseDir string) (*DiskStore, error) {
	if baseDir == "" {
		baseDir = "./media"
	}
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("resolve base dir: %w", err)
	}
	if err := os.MkdirAll(absBase, 0755); err != nil {
		return nil, fmt.Errorf("create media dir: %w", err)
	}
	return &DiskStore{baseDir: absBase}, nil
}

func (d *DiskStore) resolvePath(key string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("empty object key")
	}

	if filepath.IsAbs(key) || strings.HasPrefix(key, "/") || strings.HasPrefix(key, "\\") {
		return "", fmt.Errorf("absolute path or leading slash rejected: %s", key)
	}

	// Reject any path element that is '.' or '..'
	for _, part := range strings.FieldsFunc(key, func(r rune) bool { return r == '/' || r == '\\' }) {
		if part == ".." || part == "." {
			return "", fmt.Errorf("directory traversal rejected: %s", key)
		}
	}

	cleaned := filepath.Clean(key)
	target := filepath.Join(d.baseDir, filepath.FromSlash(cleaned))
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("resolve target path: %w", err)
	}

	rel, err := filepath.Rel(d.baseDir, targetAbs)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return "", fmt.Errorf("path outside base directory or root: %s", key)
	}

	return targetAbs, nil
}

func (d *DiskStore) Put(ctx context.Context, key, mimeType string, data []byte) error {
	path, err := d.resolvePath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (d *DiskStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	path, err := d.resolvePath(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (d *DiskStore) Delete(ctx context.Context, key string) error {
	path, err := d.resolvePath(key)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
