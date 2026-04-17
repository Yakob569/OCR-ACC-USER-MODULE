package services

import (
	"context"
	"fmt"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
)

type userService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) ports.UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) Register(ctx context.Context, email, fullName, phone, role string) (*domain.User, error) {
	// Business Logic: Check if user already exists
	existing, _ := s.repo.GetUserByEmail(ctx, email)
	if existing != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	user := &domain.User{
		Email:        email,
		FullName:     fullName,
		Phone:        &phone,
		Role:         role,
		AuthProvider: "email",
	}

	return s.repo.CreateUser(ctx, user)
}

func (s *userService) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}
