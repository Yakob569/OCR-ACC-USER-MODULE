package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cashflow/auth-service/internal/config"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioObjectStorageService struct {
	client   *minio.Client
	endpoint string
	bucket   string
	useSSL   bool
}

type disabledObjectStorageService struct {
	reason string
}

func NewObjectStorageService(cfg config.MinIOConfig) (ports.ObjectStorageService, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return &disabledObjectStorageService{reason: "MINIO_END_POINT is not configured"}, nil
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &minioObjectStorageService{
		client:   client,
		endpoint: cfg.Endpoint,
		bucket:   cfg.BucketName,
		useSSL:   cfg.UseSSL,
	}, nil
}

func (s *minioObjectStorageService) UploadReceiptImage(ctx context.Context, userID, groupID, imageID uuid.UUID, filename, contentType string, content io.Reader, contentLength int64) (string, string, *string, error) {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return "", "", nil, err
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return "", "", nil, err
		}
	}

	objectKey := buildObjectKey(userID, groupID, imageID, filename)
	_, err = s.client.PutObject(ctx, s.bucket, objectKey, content, contentLength, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", nil, err
	}

	objectURL := buildObjectURL(s.useSSL, s.endpoint, s.bucket, objectKey)
	return s.bucket, objectKey, &objectURL, nil
}

func (s *minioObjectStorageService) DownloadReceiptImage(ctx context.Context, bucket, objectKey string) ([]byte, error) {
	object, err := s.client.GetObject(ctx, bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	return io.ReadAll(object)
}

func (s *disabledObjectStorageService) UploadReceiptImage(ctx context.Context, userID, groupID, imageID uuid.UUID, filename, contentType string, content io.Reader, contentLength int64) (string, string, *string, error) {
	return "", "", nil, fmt.Errorf("object storage is unavailable: %s", s.reason)
}

func (s *disabledObjectStorageService) DownloadReceiptImage(ctx context.Context, bucket, objectKey string) ([]byte, error) {
	return nil, fmt.Errorf("object storage is unavailable: %s", s.reason)
}

func buildObjectKey(userID, groupID, imageID uuid.UUID, filename string) string {
	base := filepath.Base(strings.TrimSpace(filename))
	base = strings.ReplaceAll(base, " ", "_")
	if base == "." || base == "" {
		base = "upload"
	}
	return fmt.Sprintf("receipts/%s/%s/original/%s-%s", userID.String(), groupID.String(), imageID.String(), base)
}

func buildObjectURL(useSSL bool, endpoint, bucket, objectKey string) string {
	scheme := "http"
	if useSSL {
		scheme = "https"
	}

	u := url.URL{
		Scheme: scheme,
		Host:   endpoint,
		Path:   fmt.Sprintf("/%s/%s", bucket, objectKey),
	}
	return u.String()
}
