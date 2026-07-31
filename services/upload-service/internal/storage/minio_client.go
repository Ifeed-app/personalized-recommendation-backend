package storage

import (
	"backend-diplomski/services/upload-service/internal/config"
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	client *minio.Client
	bucket string
	region string
}

func NewMinioStorage(cfg *config.Config) (*MinioStorage, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}
	return &MinioStorage{
		client: client,
		bucket: cfg.S3Bucket,
		region: cfg.S3Region,
	}, nil
}

func (s *MinioStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucked existence: %w", err)
	}
	if exists {
		return nil
	}
	err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{
		Region: s.region,
	})
	if err != nil {
		return fmt.Errorf("failed to create a bucket: %w", err)
	}
	return nil
}

// Test file
func (s *MinioStorage) UploadTestFile(ctx context.Context) (string, error) {
	content := []byte("Hello from upload-service to MinIO at " + time.Now().Format(time.RFC3339))

	objectKey := fmt.Sprintf("test/upload-service-test-%d.txt", time.Now().Unix())

	reader := bytes.NewReader(content)

	_, err := s.client.PutObject(
		ctx,
		s.bucket,
		objectKey,
		reader,
		int64(len(content)),
		minio.PutObjectOptions{
			ContentType: "text/plain",
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload test file: %w", err)
	}

	return objectKey, nil
}

func (s *MinioStorage) UploadObject(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	if objectKey == "" {
		return fmt.Errorf("object key is req")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := s.client.PutObject(
		ctx,
		s.bucket,
		objectKey,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload object to MinIO: %w", err)
	}
	return nil
}

func (s *MinioStorage) Bucket() string {
	return s.bucket
}
