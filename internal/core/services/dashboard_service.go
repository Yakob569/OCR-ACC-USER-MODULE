package services

import (
	"context"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
)

type dashboardService struct {
	repo ports.DashboardRepository
}

func NewDashboardService(repo ports.DashboardRepository) ports.DashboardService {
	return &dashboardService{repo: repo}
}

func (s *dashboardService) GetSummary(ctx context.Context, userID uuid.UUID) (*domain.DashboardSummary, error) {
	return s.repo.GetSummary(ctx, userID)
}
