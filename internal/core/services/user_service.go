package services

import (
	"context"
	"fmt"
	"time"

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

func (s *userService) Register(ctx context.Context, email, password, fullName, phone string) (*domain.User, *domain.TokenPair, error) {
	// 1. Hash password
	pwdHash, err := s.auth.HashPassword(password)
	if err != nil {
		return nil, nil, fmt.Errorf("password hashing failed: %w", err)
	}

	// 2. Check if user already exists
	existing, _ := s.repo.GetUserByEmail(ctx, email)
	if existing != nil {
		return nil, nil, fmt.Errorf("user with email %s already exists", email)
	}

	// 3. Create profile in DB
	user := &domain.User{
		Email:        email,
		PasswordHash: pwdHash,
		FullName:     fullName,
		Phone:        &phone,
		Role:         "user",
		AuthProvider: "email",
	}

	registered, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 4. Generate initial token pair
	tokens, err := s.auth.GenerateTokenPair(registered)
	if err != nil {
		return registered, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 5. Store refresh token
	err = s.repo.StoreRefreshToken(ctx, registered.ID, tokens.RefreshToken, time.Now().Add(time.Hour*24*7))
	if err != nil {
		return registered, tokens, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return registered, tokens, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	// 1. Fetch user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Verify password
	err = s.auth.ComparePassword(password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	// 3. Generate tokens
	tokens, err := s.auth.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 4. Store refresh token
	err = s.repo.StoreRefreshToken(ctx, user.ID, tokens.RefreshToken, time.Now().Add(time.Hour*24*7))
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return tokens, nil
}

func (s *userService) RefreshToken(ctx context.Context, token string) (*domain.TokenPair, error) {
	// 1. Verify token in DB (and check expiry/revocation)
	userID, err := s.repo.GetRefreshToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 2. Get user profile
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// 3. Generate new pair
	tokens, err := s.auth.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 4. Revoke old token and store new one
	_ = s.repo.RevokeRefreshToken(ctx, token)
	err = s.repo.StoreRefreshToken(ctx, user.ID, tokens.RefreshToken, time.Now().Add(time.Hour*24*7))
	if err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	return tokens, nil
}

func (s *userService) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}
