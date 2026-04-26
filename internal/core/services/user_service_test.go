package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type stubUserRepo struct {
	getUserByEmailFn     func(ctx context.Context, email string) (*domain.User, error)
	createUserFn         func(ctx context.Context, user *domain.User) (*domain.User, error)
	getUserByIDFn        func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	storeRefreshTokenFn  func(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	revokeRefreshTokenFn func(ctx context.Context, token string) error
	getRefreshTokenFn    func(ctx context.Context, token string) (uuid.UUID, error)
}

func (s *stubUserRepo) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if s.createUserFn == nil {
		return nil, nil
	}
	return s.createUserFn(ctx, user)
}

func (s *stubUserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if s.getUserByEmailFn == nil {
		return nil, nil
	}
	return s.getUserByEmailFn(ctx, email)
}

func (s *stubUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if s.getUserByIDFn == nil {
		return nil, nil
	}
	return s.getUserByIDFn(ctx, id)
}

func (s *stubUserRepo) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	if s.storeRefreshTokenFn == nil {
		return nil
	}
	return s.storeRefreshTokenFn(ctx, userID, token, expiresAt)
}

func (s *stubUserRepo) RevokeRefreshToken(ctx context.Context, token string) error {
	if s.revokeRefreshTokenFn == nil {
		return nil
	}
	return s.revokeRefreshTokenFn(ctx, token)
}

func (s *stubUserRepo) GetRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	if s.getRefreshTokenFn == nil {
		return uuid.Nil, nil
	}
	return s.getRefreshTokenFn(ctx, token)
}

type stubAuthService struct {
	hashPasswordFn      func(password string) (string, error)
	comparePasswordFn   func(password, hash string) error
	generateTokenPairFn func(user *domain.User) (*domain.TokenPair, error)
	validateTokenFn     func(token string) (uuid.UUID, error)
	refreshTokenFn      func(token string) (*domain.TokenPair, error)
}

func (s *stubAuthService) HashPassword(password string) (string, error) {
	if s.hashPasswordFn == nil {
		return "", nil
	}
	return s.hashPasswordFn(password)
}

func (s *stubAuthService) ComparePassword(password, hash string) error {
	if s.comparePasswordFn == nil {
		return nil
	}
	return s.comparePasswordFn(password, hash)
}

func (s *stubAuthService) GenerateTokenPair(user *domain.User) (*domain.TokenPair, error) {
	if s.generateTokenPairFn == nil {
		return nil, nil
	}
	return s.generateTokenPairFn(user)
}

func (s *stubAuthService) ValidateToken(token string) (uuid.UUID, error) {
	if s.validateTokenFn == nil {
		return uuid.Nil, nil
	}
	return s.validateTokenFn(token)
}

func (s *stubAuthService) RefreshToken(token string) (*domain.TokenPair, error) {
	if s.refreshTokenFn == nil {
		return nil, nil
	}
	return s.refreshTokenFn(token)
}

func TestRegisterReturnsLookupError(t *testing.T) {
	lookupErr := errors.New("database unavailable")
	repo := &stubUserRepo{
		getUserByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, lookupErr
		},
		createUserFn: func(ctx context.Context, user *domain.User) (*domain.User, error) {
			t.Fatal("CreateUser should not be called when existence lookup fails")
			return nil, nil
		},
	}
	auth := &stubAuthService{
		hashPasswordFn: func(password string) (string, error) { return "hash", nil },
	}

	svc := NewUserService(repo, auth)

	_, _, err := svc.Register(context.Background(), "user@example.com", "secret", "User", "")
	if err == nil {
		t.Fatal("expected register to fail when lookup errors")
	}
	if !errors.Is(err, lookupErr) {
		t.Fatalf("expected error to wrap lookup error, got %v", err)
	}
}

func TestRegisterIgnoresNotFoundLookup(t *testing.T) {
	repo := &stubUserRepo{
		getUserByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, pgx.ErrNoRows
		},
		createUserFn: func(ctx context.Context, user *domain.User) (*domain.User, error) {
			return &domain.User{ID: uuid.New(), Email: user.Email}, nil
		},
		storeRefreshTokenFn: func(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
			return nil
		},
	}
	auth := &stubAuthService{
		hashPasswordFn: func(password string) (string, error) { return "hash", nil },
		generateTokenPairFn: func(user *domain.User) (*domain.TokenPair, error) {
			return &domain.TokenPair{
				AccessToken:  "access",
				RefreshToken: "refresh",
				ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			}, nil
		},
	}

	svc := NewUserService(repo, auth)

	user, tokens, err := svc.Register(context.Background(), "user@example.com", "secret", "User", "")
	if err != nil {
		t.Fatalf("expected register to succeed, got %v", err)
	}
	if user == nil || tokens == nil {
		t.Fatal("expected user and tokens to be returned")
	}
}
