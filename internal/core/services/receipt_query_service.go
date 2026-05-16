package services

import (
	"context"
	"fmt"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
)

type receiptQueryService struct {
	groupRepo      ports.ReceiptGroupRepository
	imageRepo      ports.ReceiptImageRepository
	extractionRepo ports.OCRExtractionRepository
}

func NewReceiptQueryService(
	groupRepo ports.ReceiptGroupRepository,
	imageRepo ports.ReceiptImageRepository,
	extractionRepo ports.OCRExtractionRepository,
) ports.ReceiptQueryService {
	return &receiptQueryService{
		groupRepo:      groupRepo,
		imageRepo:      imageRepo,
		extractionRepo: extractionRepo,
	}
}

func (s *receiptQueryService) ListGroupImages(ctx context.Context, userID, groupID uuid.UUID, limit, offset int) ([]domain.ReceiptImage, error) {
	if _, err := s.groupRepo.GetByUserAndID(ctx, userID, groupID); err != nil {
		return nil, fmt.Errorf("group not found")
	}
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.imageRepo.ListByGroup(ctx, userID, groupID, limit, offset)
}

func (s *receiptQueryService) GetImage(ctx context.Context, userID, imageID uuid.UUID) (*domain.ReceiptImage, error) {
	return s.imageRepo.GetByUserAndID(ctx, userID, imageID)
}

func (s *receiptQueryService) DeleteImage(ctx context.Context, userID, imageID uuid.UUID) error {
	if _, err := s.imageRepo.GetByUserAndID(ctx, userID, imageID); err != nil {
		return err
	}
	return s.imageRepo.TrashImage(ctx, imageID)
}


func (s *receiptQueryService) GetImageResult(ctx context.Context, userID, imageID uuid.UUID) (*domain.OCRExtraction, error) {
	if _, err := s.imageRepo.GetByUserAndID(ctx, userID, imageID); err != nil {
		return nil, err
	}
	return s.extractionRepo.GetByReceiptImageID(ctx, imageID)
}

func (s *receiptQueryService) ListGroupResults(ctx context.Context, userID, groupID uuid.UUID, limit, offset int) ([]domain.OCRExtraction, error) {
	if _, err := s.groupRepo.GetByUserAndID(ctx, userID, groupID); err != nil {
		return nil, fmt.Errorf("group not found")
	}
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.extractionRepo.ListByGroup(ctx, userID, groupID, limit, offset)
}
