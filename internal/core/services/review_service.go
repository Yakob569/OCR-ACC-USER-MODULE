package services

import (
	"context"
	"fmt"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
)

type receiptReviewService struct {
	reviewRepo ports.ReceiptReviewRepository
	imageRepo  ports.ReceiptImageRepository
	groupRepo  ports.ReceiptGroupRepository
}

func NewReceiptReviewService(
	reviewRepo ports.ReceiptReviewRepository,
	imageRepo ports.ReceiptImageRepository,
	groupRepo ports.ReceiptGroupRepository,
) ports.ReceiptReviewService {
	return &receiptReviewService{
		reviewRepo: reviewRepo,
		imageRepo:  imageRepo,
		groupRepo:  groupRepo,
	}
}

func (s *receiptReviewService) SubmitReview(ctx context.Context, input domain.SubmitReceiptReviewInput) (*domain.ReceiptReview, error) {
	image, err := s.imageRepo.GetByUserAndID(ctx, input.ReviewedByUserID, input.ReceiptImageID)
	if err != nil {
		return nil, fmt.Errorf("image not found")
	}
	if input.QualityLabel != "accurate" && input.QualityLabel != "partially_accurate" && input.QualityLabel != "inaccurate" {
		return nil, fmt.Errorf("invalid quality label")
	}

	_, existingErr := s.reviewRepo.GetByReceiptImageID(ctx, input.ReceiptImageID)
	review, err := s.reviewRepo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	reviewStatus := domain.ReviewStatusReviewed
	if input.IsAccepted {
		reviewStatus = domain.ReviewStatusAccepted
	} else if input.QualityLabel == "inaccurate" {
		reviewStatus = domain.ReviewStatusRejected
	}

	if err := s.imageRepo.UpdateStatuses(ctx, image.ID, image.UploadStatus, image.OCRStatus, reviewStatus); err != nil {
		return nil, err
	}

	if existingErr != nil {
		_ = s.groupRepo.IncrementImageCounters(ctx, image.GroupID, 0, 0, 0, 0, 0, 1, 0)
	}

	return review, nil
}
