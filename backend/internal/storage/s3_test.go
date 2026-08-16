package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3Client struct {
	putFn    func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	getFn    func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	deleteFn func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.putFn != nil {
		return m.putFn(ctx, params, optFns...)
	}
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.getFn != nil {
		return m.getFn(ctx, params, optFns...)
	}
	return &s3.GetObjectOutput{}, nil
}

func (m *mockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, params, optFns...)
	}
	return &s3.DeleteObjectOutput{}, nil
}

func TestS3StorageOperations(t *testing.T) {
	var capturedPut *s3.PutObjectInput
	var capturedGet *s3.GetObjectInput
	var capturedDelete *s3.DeleteObjectInput

	mockClient := &mockS3Client{
		putFn: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			capturedPut = params
			return &s3.PutObjectOutput{}, nil
		},
		getFn: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			capturedGet = params
			return &s3.GetObjectOutput{
				Body: io.NopCloser(bytes.NewReader([]byte("s3 content"))),
			}, nil
		},
		deleteFn: func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
			capturedDelete = params
			return &s3.DeleteObjectOutput{}, nil
		},
	}

	store := NewS3Storage(mockClient, "test-bucket")
	ctx := context.Background()

	// Put
	data := []byte("hello s3")
	if err := store.Put(ctx, "media/file.txt", "text/plain", data); err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if capturedPut == nil || *capturedPut.Bucket != "test-bucket" || *capturedPut.Key != "media/file.txt" {
		t.Fatalf("unexpected put params: %+v", capturedPut)
	}

	// Open
	rc, err := store.Open(ctx, "media/file.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer rc.Close()
	content, _ := io.ReadAll(rc)
	if string(content) != "s3 content" {
		t.Fatalf("expected 's3 content', got %q", string(content))
	}
	if capturedGet == nil || *capturedGet.Bucket != "test-bucket" || *capturedGet.Key != "media/file.txt" {
		t.Fatalf("unexpected get params: %+v", capturedGet)
	}

	// Delete
	if err := store.Delete(ctx, "media/file.txt"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if capturedDelete == nil || *capturedDelete.Bucket != "test-bucket" || *capturedDelete.Key != "media/file.txt" {
		t.Fatalf("unexpected delete params: %+v", capturedDelete)
	}
}

func TestS3StorageErrors(t *testing.T) {
	mockClient := &mockS3Client{
		getFn: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return nil, errors.New("NoSuchKey")
		},
	}

	store := NewS3Storage(mockClient, "test-bucket")
	_, err := store.Open(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected Open error, got nil")
	}
}
