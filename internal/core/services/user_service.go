package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	log.Printf("➡️  [UserService.Register] Attempting to register user with email: %s", email)

	// 1. Hash password
	log.Printf("ℹ️  [UserService.Register] Hashing password for: %s", email)
	pwdHash, err := s.auth.HashPassword(password)
	if err != nil {
		log.Printf("❌ [UserService.Register] Password hashing failed for %s: %v", email, err)
		return nil, nil, fmt.Errorf("password hashing failed: %w", err)
	}
	log.Printf("✅ [UserService.Register] Password hashed successfully for: %s", email)


	// 2. Check if user already exists
	log.Printf("ℹ️  [UserService.Register] Checking if user exists with email: %s", email)
	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("❌ [UserService.Register] Failed to check existing user for %s: %v", email, err)
		return nil, nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existing != nil {
		log.Printf("❌ [UserService.Register] User with email %s already exists", email)
		return nil, nil, fmt.Errorf("user with email %s already exists", email)
	}
	log.Printf("✅ [UserService.Register] No existing user found with email: %s", email)


	// 3. Create profile in DB
	user := &domain.User{
		Email:        email,
		PasswordHash: pwdHash,
		FullName:     fullName,
		Phone:        &phone,
		Role:         "user",
		AuthProvider: "email",
	}

	log.Printf("ℹ️  [UserService.Register] Calling repository to create user for email: %s", email)
	registered, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		log.Printf("❌ [UserService.Register] Failed to create user in repository for %s: %v", email, err)
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}
	log.Printf("✅ [UserService.Register] User created in repository with ID: %s", registered.ID)


	// 4. Generate initial token pair
	log.Printf("ℹ️  [UserService.Register] Generating tokens for user ID: %s", registered.ID)
	tokens, err := s.auth.GenerateTokenPair(registered)
	if err != nil {
		log.Printf("❌ [UserService.Register] Failed to generate tokens for user ID %s: %v", registered.ID, err)
		return registered, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}
	log.Printf("✅ [UserService.Register] Tokens generated successfully for user ID: %s", registered.ID)


	// 5. Store refresh token
	log.Printf("ℹ️  [UserService.Register] Storing refresh token for user ID: %s", registered.ID)
	err = s.repo.StoreRefreshToken(ctx, registered.ID, tokens.RefreshToken, time.Now().Add(time.Hour*24*7))
	if err != nil {
		log.Printf("❌ [UserService.Register] Failed to store refresh token for user ID %s: %v", registered.ID, err)
		return registered, tokens, fmt.Errorf("failed to store refresh token: %w", err)
	}
	log.Printf("✅ [UserService.Register] Refresh token stored successfully for user ID: %s", registered.ID)


	log.Printf("⬅️  [UserService.Register] Registration complete for user ID: %s, email: %s", registered.ID, email)
	return registered, tokens, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	log.Printf("➡️  [UserService.Login] Attempting login for email: %s", email)

	// 1. Fetch user by email
	log.Printf("ℹ️  [UserService.Login] Fetching user by email: %s", email)
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		log.Printf("❌ [UserService.Login] User not found for email %s: %v", email, err)
		return nil, fmt.Errorf("user not found: %w", err)
	}
	log.Printf("✅ [UserService.Login] User found with ID: %s for email: %s", user.ID, email)


	// 2. Verify password
	log.Printf("ℹ️  [UserService.Login] Verifying password for user ID: %s", user.ID)
	err = s.auth.ComparePassword(password, user.PasswordHash)
	if err != nil {
		log.Printf("❌ [UserService.Login] Invalid password for user ID %s: %v", user.ID, err)
		return nil, fmt.Errorf("invalid password")
	}
	log.Printf("✅ [UserService.Login] Password verified for user ID: %s", user.ID)


	// 3. Generate tokens
	log.Printf("ℹ️  [UserService.Login] Generating tokens for user ID: %s", user.ID)
	tokens, err := s.auth.GenerateTokenPair(user)
	if err != nil {
		log.Printf("❌ [UserService.Login] Failed to generate tokens for user ID %s: %v", user.ID, err)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}
	log.Printf("✅ [UserService.Login] Tokens generated successfully for user ID: %s", user.ID)


	// 4. Store refresh token
	log.Printf("ℹ️  [UserService.Login] Storing refresh token for user ID: %s", user.ID)
	err = s.repo.StoreRefreshToken(ctx, user.ID, tokens.RefreshToken, time.Now().Add(time.Hour*24*7))
	if err != nil {
		log.Printf("❌ [UserService.Login] Failed to store refresh token for user ID %s: %v", user.ID, err)
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}
	log.Printf("✅ [UserService.Login] Refresh token stored successfully for user ID: %s", user.ID)


	log.Printf("⬅️  [UserService.Login] Login successful for email: %s", email)
	return tokens, nil
}

func (s *userService) RefreshToken(ctx context.Context, token string) (*domain.TokenPair, error) {
	log.Printf("➡️  [UserService.RefreshToken] Attempting token refresh")

	// 1. Verify token in DB (and check expiry/revocation)
	log.Printf("ℹ️  [UserService.RefreshToken] Validating refresh token in repository")
	userID, err := s.repo.GetRefreshToken(ctx, token)
	if err != nil {
		log.Printf("❌ [UserService.RefreshToken] Invalid refresh token: %v", err)
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}
	log.Printf("✅ [UserService.RefreshToken] Refresh token validated for user ID: %s", userID)


	// 2. Get user profile
	log.Printf("ℹ️  [UserService.RefreshToken] Fetching user profile for ID: %s", userID)
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("❌ [UserService.RefreshToken] Failed to fetch user for ID %s: %v", userID, err)
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	log.Printf("✅ [UserService.RefreshToken] User profile fetched for ID: %s, email: %s", userID, user.Email)


	// 3. Generate new pair
	log.Printf("ℹ️  [UserService.RefreshToken] Generating new token pair for user ID: %s", userID)
	tokens, err := s.auth.GenerateTokenPair(user)
	if err != nil {
		log.Printf("❌ [UserService.RefreshToken] Failed to generate tokens for user ID %s: %v", userID, err)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}
	log.Printf("✅ [UserService.RefreshToken] New token pair generated for user ID: %s", userID)


	// 4. Revoke old token and store new one
	log.Printf("ℹ️  [UserService.RefreshToken] Revoking old refresh token and storing new one for user ID: %s", userID)
	_ = s.repo.RevokeRefreshToken(ctx, token) // Log inside RevokeRefreshToken
	err = s.repo.StoreRefreshToken(ctx, user.ID, tokens.RefreshToken, time.Now().Add(time.Hour*24*7)) // Log inside StoreRefreshToken
	if err != nil {
		log.Printf("❌ [UserService.RefreshToken] Failed to store new refresh token for user ID %s: %v", userID, err)
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}
	log.Printf("✅ [UserService.RefreshToken] Old refresh token revoked and new one stored for user ID: %s", userID)


	log.Printf("⬅️  [UserService.RefreshToken] Token refresh complete for user ID: %s", userID)
	return tokens, nil
}

func (s *userService) Logout(ctx context.Context, token string) error {
	log.Printf("➡️  [UserService.Logout] Attempting logout")

	if token == "" {
		log.Printf("❌ [UserService.Logout] Refresh token is required for logout")
		return fmt.Errorf("refresh token is required")
	}
	log.Printf("ℹ️  [UserService.Logout] Refresh token received (truncated): %s...", token[:10])


	log.Printf("ℹ️  [UserService.Logout] Validating refresh token in repository")
	if _, err := s.repo.GetRefreshToken(ctx, token); err != nil {
		log.Printf("❌ [UserService.Logout] Invalid refresh token during validation: %v", err)
		return fmt.Errorf("invalid refresh token: %w", err)
	}
	log.Printf("✅ [UserService.Logout] Refresh token validated successfully")


	log.Printf("ℹ️  [UserService.Logout] Revoking refresh token in repository")
	if err := s.repo.RevokeRefreshToken(ctx, token); err != nil {
		log.Printf("❌ [UserService.Logout] Failed to revoke refresh token: %v", err)
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	log.Printf("✅ [UserService.Logout] Refresh token revoked successfully")


	log.Printf("⬅️  [UserService.Logout] Logout complete")
	return nil
}

func (s *userService) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	log.Printf("➡️  [UserService.GetProfile] Attempting to fetch profile for user ID: %s", id)

	log.Printf("ℹ️  [UserService.GetProfile] Calling repository to get user by ID: %s", id)
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		log.Printf("❌ [UserService.GetProfile] Failed to fetch user from repository for ID %s: %v", id, err)
		return nil, err
	}
	log.Printf("✅ [UserService.GetProfile] User profile fetched successfully for ID: %s, email: %s", id, user.Email)

	log.Printf("⬅️  [UserService.GetProfile] Profile retrieval complete for user ID: %s", id)
	return user, nil
}
