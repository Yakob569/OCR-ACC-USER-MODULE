package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
)

type receiptUploadService struct {
	groupRepo     ports.ReceiptGroupRepository
	imageRepo     ports.ReceiptImageRepository
	jobRepo       ports.OCRJobRepository
	objectStorage ports.ObjectStorageService
	maxFiles      int
	maxFileSizeMB int
}

func NewReceiptUploadService(
	groupRepo ports.ReceiptGroupRepository,
	imageRepo ports.ReceiptImageRepository,
	jobRepo ports.OCRJobRepository,
	objectStorage ports.ObjectStorageService,
	maxFiles int,
	maxFileSizeMB int,
) ports.ReceiptUploadService {
	return &receiptUploadService{
		groupRepo:     groupRepo,
		imageRepo:     imageRepo,
		jobRepo:       jobRepo,
		objectStorage: objectStorage,
		maxFiles:      maxFiles,
		maxFileSizeMB: maxFileSizeMB,
	}
}

func (s *receiptUploadService) UploadGroupImages(ctx context.Context, groupID, userID uuid.UUID, files []ports.ReceiptFile) (*ports.ReceiptUploadResult, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}
	if groupID == uuid.Nil {
		return nil, fmt.Errorf("group ID is required")
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("at least one image is required")
	}
	if len(files) > s.maxFiles {
		return nil, fmt.Errorf("too many files: maximum is %d", s.maxFiles)
	}

	if _, err := s.groupRepo.GetByUserAndID(ctx, userID, groupID); err != nil {
		return nil, fmt.Errorf("group not found")
	}

	maxBytes := int64(s.maxFileSizeMB) * 1024 * 1024
	inputs := make([]domain.ReceiptImageUploadInput, 0, len(files))

	for _, file := range files {
		if !strings.HasPrefix(file.ContentType, "image/") {
			return nil, fmt.Errorf("only image uploads are supported")
		}
		if len(file.Bytes) == 0 {
			return nil, fmt.Errorf("uploaded file is empty")
		}
		if int64(len(file.Bytes)) > maxBytes {
			return nil, fmt.Errorf("file %s exceeds the %d MB upload limit", file.Filename, s.maxFileSizeMB)
		}

		imageID := uuid.New()
		checksum := sha256.Sum256(file.Bytes)
		bucket, objectKey, objectURL, err := s.objectStorage.UploadReceiptImage(
			ctx,
			userID,
			groupID,
			imageID,
			file.Filename,
			file.ContentType,
			bytes.NewReader(file.Bytes),
			int64(len(file.Bytes)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to upload %s: %w", file.Filename, err)
		}

		var storageURL string
		if objectURL != nil {
			storageURL = *objectURL
		}

		inputs = append(inputs, domain.ReceiptImageUploadInput{
			ID:               imageID,
			GroupID:          groupID,
			UserID:           userID,
			OriginalFilename: file.Filename,
			MIMEType:         file.ContentType,
			FileSizeBytes:    int64(len(file.Bytes)),
			ChecksumSHA256:   hex.EncodeToString(checksum[:]),
			StorageBucket:    bucket,
			StorageObjectKey: objectKey,
			StorageURL:       storageURL,
		})
	}

	images, err := s.imageRepo.CreateMany(ctx, inputs)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	jobInputs := make([]ports.OCRJobCreateInput, 0, len(images))
	for _, image := range images {
		jobInputs = append(jobInputs, ports.OCRJobCreateInput{
			ReceiptImageID: image.ID,
			GroupID:        image.GroupID,
			UserID:         image.UserID,
			Status:         domain.JobStatusQueued,
			AttemptCount:   0,
			MaxAttempts:    3,
			QueuedAt:       now,
		})
	}

	jobs, err := s.jobRepo.CreateMany(ctx, jobInputs)
	if err != nil {
		return nil, err
	}

	if err := s.groupRepo.IncrementImageCounters(ctx, groupID, len(images), len(images), 0, 0, 0, 0, 0); err != nil {
		return nil, err
	}
	if err := s.groupRepo.UpdateStatus(ctx, groupID, domain.GroupStatusQueued); err != nil {
		return nil, err
	}

	return &ports.ReceiptUploadResult{
		Images: images,
		Jobs:   jobs,
	}, nil
}
