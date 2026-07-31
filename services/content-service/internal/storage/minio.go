package storage

import (
	"backend-diplomski/services/content-service/internal/config"
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	client       *minio.Client
	publicClient *minio.Client
	bucket       string
}

func NewMinioStorage(cfg *config.Config) (*MinioStorage, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
		Region: cfg.S3Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}
	publicEndpoint := cfg.S3PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.S3Endpoint
	}
	publicClient, err := minio.New(publicEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
		Region: cfg.S3Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create public minio client: %w", err)
	}
	return &MinioStorage{
		client:       client,
		publicClient: publicClient,
		bucket:       cfg.S3Bucket,
	}, nil
}

func (s *MinioStorage) PresignedObject(ctx context.Context, objectKey string) (string, error) {
	if objectKey == "" {
		return "", fmt.Errorf("object key is required")
	}
	presignedURL, err := s.publicClient.PresignedGetObject(
		ctx,
		s.bucket,
		objectKey,
		1*time.Hour,
		url.Values{},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create presigned get object url: %w", err)
	}
	return presignedURL.String(), nil
}

func (s *MinioStorage) Bucket() string {
	return s.bucket
}
