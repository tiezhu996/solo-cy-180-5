package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/oralhistory/oralhistory/internal/config"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/util"
)

// StorageService 录音文件对象存储服务（MinIO）。
type StorageService interface {
	Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error
	Get(ctx context.Context, objectKey string) (*minio.Object, error)
	Remove(ctx context.Context, objectKey string) error
}

type storageService struct {
	client *minio.Client
	bucket string
	logger *slog.Logger
}

// NewStorageService 构造对象存储服务。
func NewStorageService(cfg *config.Config, logger *slog.Logger) (StorageService, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinIOBucket)
	if err != nil {
		return nil, fmt.Errorf("check minio bucket %s: %w", cfg.MinIOBucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinIOBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("make minio bucket %s: %w", cfg.MinIOBucket, err)
		}
	}
	logger.Info("minio connected", "endpoint", cfg.MinIOEndpoint, "bucket", cfg.MinIOBucket)
	return &storageService{client: client, bucket: cfg.MinIOBucket, logger: logger}, nil
}

func (s *storageService) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("upload object %s: %w", objectKey, err)
	}
	return nil
}

func (s *storageService) Get(ctx context.Context, objectKey string) (*minio.Object, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object %s: %w", objectKey, err)
	}
	return obj, nil
}

func (s *storageService) Remove(ctx context.Context, objectKey string) error {
	if objectKey == "" {
		return util.NewAppError(constants.CodeBadRequest, "录音对象 key 为空", nil)
	}
	if err := s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("remove object %s: %w", objectKey, err)
	}
	return nil
}
