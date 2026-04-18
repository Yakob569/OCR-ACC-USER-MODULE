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
	auth ports.AuthService
}

func NewUserService(repo ports.UserRepository, auth ports.AuthService) ports.UserService {
	return &userService{
		repo: repo,
		auth: auth,
	}
}

func (s *userService) Register(ctx context.Context, email, password, fullName, phone, role string) (*domain.User, error) {
	// 1. Create in Supabase Auth
	_, err := s.auth.SignUp(ctx, email, password)
	if err != nil {
		return nil, fmt.Errorf("auth signup failed: %w", err)
	}

	// 2. Check if user already exists in local DB
	existing, _ := s.repo.GetUserByEmail(ctx, email)
	if existing != nil {
		return existing, nil
	}

	// 3. Create profile in local DB
	user := &domain.User{
		Email:        email,
		FullName:     fullName,
		Phone:        &phone,
		Role:         role,
		AuthProvider: "email",
	}

	return s.repo.CreateUser(ctx, user)
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	return s.auth.Login(ctx, email, password)
}

func (s *userService) ResetPassword(ctx context.Context, email string) error {
	return s.auth.ResetPassword(ctx, email)
}

func (s *userService) GetSocialLoginURL(provider string) (string, error) {
	return s.auth.GetSocialLoginURL(provider)
}

func (s *userService) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}
