package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3ClientAPI defines the subset of the AWS S3 client needed by S3Storage.
type S3ClientAPI interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// S3Storage provides S3-backed media storage satisfying the MediaStore interface.
type S3Storage struct {
	client S3ClientAPI
	bucket string
}

func NewS3Storage(client S3ClientAPI, bucket string) *S3Storage {
	return &S3Storage{client: client, bucket: bucket}
}

func (s *S3Storage) Put(ctx context.Context, key, mimeType string, data []byte) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	}
	if mimeType != "" {
		input.ContentType = aws.String(mimeType)
	}
	_, err := s.client.PutObject(ctx, input)
	return err
}

func (s *S3Storage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// Upload preserves compatibility with the legacy upload interface.
func (s *S3Storage) Upload(data []byte, key, mimeType string) (string, error) {
	if err := s.Put(context.Background(), key, mimeType, data); err != nil {
		return "", err
	}
	return fmt.Sprintf("s3://%s/%s", s.bucket, key), nil
}
