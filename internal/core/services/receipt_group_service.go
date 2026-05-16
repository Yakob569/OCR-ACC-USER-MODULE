package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
)

type receiptGroupService struct {
	repo ports.ReceiptGroupRepository
}

func NewReceiptGroupService(repo ports.ReceiptGroupRepository) ports.ReceiptGroupService {
	return &receiptGroupService{repo: repo}
}

func (s *receiptGroupService) CreateGroup(ctx context.Context, input domain.CreateReceiptGroupInput) (*domain.ReceiptGroup, error) {
	name := strings.TrimSpace(input.Name)
	if input.UserID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}
	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	input.Name = name
	return s.repo.Create(ctx, input)
}

func (s *receiptGroupService) ListGroups(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.ReceiptGroup, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListByUser(ctx, userID, limit, offset)
}

func (s *receiptGroupService) GetGroup(ctx context.Context, userID, groupID uuid.UUID) (*domain.ReceiptGroup, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}
	if groupID == uuid.Nil {
		return nil, fmt.Errorf("group ID is required")
	}

	return s.repo.GetByUserAndID(ctx, userID, groupID)
}

func (s *receiptGroupService) DeleteGroup(ctx context.Context, userID, groupID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if groupID == uuid.Nil {
		return fmt.Errorf("group ID is required")
	}

	// Verify existence and ownership
	_, err := s.repo.GetByUserAndID(ctx, userID, groupID)
	if err != nil {
		return err
	}

	return s.repo.SoftDelete(ctx, groupID)
}

