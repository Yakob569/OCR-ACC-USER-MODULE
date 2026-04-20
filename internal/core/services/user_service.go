package services

import (
	"context"
	"fmt"
	"log"
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
	log.Printf("[UserService] Starting registration for: %s", email)

	// 1. Hash password
	log.Printf("[UserService] Hashing password for: %s", email)
	pwdHash, err := s.auth.HashPassword(password)
	if err != nil {
		return nil, nil, fmt.Errorf("password hashing failed: %w", err)
	}

	// 2. Check if user already exists
	log.Printf("[UserService] Checking if user exists: %s", email)
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

	log.Printf("[UserService] Creating user in repository: %s", email)
	registered, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 4. Generate initial token pair
	log.Printf("[UserService] Generating tokens for user: %s", registered.ID)
	tokens, err := s.auth.GenerateTokenPair(registered)
	if err != nil {
		return registered, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 5. Store refresh token
	log.Printf("[UserService] Storing refresh token for user: %s", registered.ID)
	err = s.repo.StoreRefreshToken(ctx, registered.ID, tokens.RefreshToken, time.Now().Add(time.Hour*24*7))
	if err != nil {
		return registered, tokens, fmt.Errorf("failed to store refresh token: %w", err)
	}

	log.Printf("[UserService] Registration complete for: %s", email)
	return registered, tokens, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	log.Printf("[UserService] Starting login for: %s", email)

	// 1. Fetch user by email
	log.Printf("[UserService] Fetching user: %s", email)
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Verify password
	log.Printf("[UserService] Verifying password for: %s", email)
	err = s.auth.ComparePassword(password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	// 3. Generate tokens
	log.Printf("[UserService] Generating tokens for user: %s", user.ID)
	tokens, err := s.auth.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 4. Store refresh token
	log.Printf("[UserService] Storing refresh token for user: %s", user.ID)
	err = s.repo.StoreRefreshToken(ctx, user.ID, tokens.RefreshToken, time.Now().Add(time.Hour*24*7))
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	log.Printf("[UserService] Login successful for: %s", email)
	return tokens, nil
}

func (s *userService) RefreshToken(ctx context.Context, token string) (*domain.TokenPair, error) {
	log.Printf("[UserService] Starting token refresh")

	// 1. Verify token in DB (and check expiry/revocation)
	log.Printf("[UserService] Validating refresh token in repository")
	userID, err := s.repo.GetRefreshToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 2. Get user profile
	log.Printf("[UserService] Fetching user profile for ID: %s", userID)
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// 3. Generate new pair
	log.Printf("[UserService] Generating new token pair for user: %s", userID)
	tokens, err := s.auth.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 4. Revoke old token and store new one
	log.Printf("[UserService] Revoking old refresh token and storing new one")
	_ = s.repo.RevokeRefreshToken(ctx, token)
	err = s.repo.StoreRefreshToken(ctx, user.ID, tokens.RefreshToken, time.Now().Add(time.Hour*24*7))
	if err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	log.Printf("[UserService] Token refresh complete")
	return tokens, nil
}

func (s *userService) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	log.Printf("[UserService] Fetching profile for user ID: %s", id)
	return s.repo.GetUserByID(ctx, id)
}
