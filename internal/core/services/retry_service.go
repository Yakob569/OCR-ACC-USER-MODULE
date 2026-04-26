package services

import (
	"context"
	"fmt"
	"time"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
)

type ocrRetryService struct {
	imageRepo ports.ReceiptImageRepository
	jobRepo   ports.OCRJobRepository
	groupRepo ports.ReceiptGroupRepository
}

func NewOCRRetryService(
	imageRepo ports.ReceiptImageRepository,
	jobRepo ports.OCRJobRepository,
	groupRepo ports.ReceiptGroupRepository,
) ports.OCRRetryService {
	return &ocrRetryService{
		imageRepo: imageRepo,
		jobRepo:   jobRepo,
		groupRepo: groupRepo,
	}
}

func (s *ocrRetryService) RetryImage(ctx context.Context, userID, imageID uuid.UUID) (*domain.OCRJob, error) {
	image, err := s.imageRepo.GetByUserAndID(ctx, userID, imageID)
	if err != nil {
		return nil, fmt.Errorf("image not found")
	}

	if image.OCRStatus != domain.OCRStatusFailed && image.OCRStatus != domain.OCRStatusNeedsReview {
		return nil, fmt.Errorf("image is not eligible for retry")
	}

	if err := s.imageRepo.UpdateProcessingResult(ctx, image.ID, domain.OCRStatusQueued, nil, nil, nil, nil); err != nil {
		return nil, err
	}

	jobs, err := s.jobRepo.CreateMany(ctx, []ports.OCRJobCreateInput{
		{
			ReceiptImageID: image.ID,
			GroupID:        image.GroupID,
			UserID:         image.UserID,
			Status:         domain.JobStatusQueued,
			AttemptCount:   0,
			MaxAttempts:    3,
			QueuedAt:       time.Now().UTC(),
		},
	})
	if err != nil {
		return nil, err
	}

	if _, err := s.groupRepo.RefreshAggregateState(ctx, image.GroupID); err != nil {
		return nil, err
	}

	return &jobs[0], nil
}
