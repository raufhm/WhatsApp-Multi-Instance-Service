package storage

import (
	"context"
	"io"
	"os"
	"testing"
)

func TestDiskStoreRoundTrip(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "diskstore_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewDiskStore(tempDir)
	if err != nil {
		t.Fatalf("NewDiskStore failed: %v", err)
	}

	ctx := context.Background()
	key := "media/test-file.txt"
	content := []byte("hello media store")

	// Put
	if err := store.Put(ctx, key, "text/plain", content); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Open
	rc, err := store.Open(ctx, key)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer rc.Close()

	readBytes, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(readBytes) != string(content) {
		t.Fatalf("expected %q, got %q", string(content), string(readBytes))
	}

	// Delete
	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Open after delete should fail
	_, err = store.Open(ctx, key)
	if err == nil {
		t.Fatal("expected Open to fail after Delete, but it succeeded")
	}
}

func TestDiskStorePathTraversal(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "diskstore_traversal_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewDiskStore(tempDir)
	if err != nil {
		t.Fatalf("NewDiskStore failed: %v", err)
	}

	ctx := context.Background()
	badKeys := []string{
		"../secret.txt",
		"media/../../etc/passwd",
		"..",
		"media/..",
		"/etc/shadow",
		"",
	}

	for _, badKey := range badKeys {
		err := store.Put(ctx, badKey, "text/plain", []byte("bad"))
		if err == nil {
			t.Errorf("expected Put with bad key %q to fail, but it succeeded", badKey)
		}

		_, err = store.Open(ctx, badKey)
		if err == nil {
			t.Errorf("expected Open with bad key %q to fail, but it succeeded", badKey)
		}

		err = store.Delete(ctx, badKey)
		if err == nil && badKey != "" {
			t.Errorf("expected Delete with bad key %q to fail, but it succeeded", badKey)
		}
	}
}

func TestDiskStoreNestedDirectories(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "diskstore_nested_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewDiskStore(tempDir)
	if err != nil {
		t.Fatalf("NewDiskStore failed: %v", err)
	}

	ctx := context.Background()
	key := "deeply/nested/sub/dir/media-item.png"
	data := []byte{0x89, 0x50, 0x4E, 0x47}

	if err := store.Put(ctx, key, "image/png", data); err != nil {
		t.Fatalf("Put into nested path failed: %v", err)
	}

	rc, err := store.Open(ctx, key)
	if err != nil {
		t.Fatalf("Open from nested path failed: %v", err)
	}
	defer rc.Close()

	readBytes, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if len(readBytes) != len(data) {
		t.Fatalf("expected length %d, got %d", len(data), len(readBytes))
	}
}
