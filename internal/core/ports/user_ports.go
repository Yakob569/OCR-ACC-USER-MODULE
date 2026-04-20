package ports

import (
	"context"
	"time"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/google/uuid"
)

// UserRepository defines the persistence contract for users
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	RevokeRefreshToken(ctx context.Context, token string) error
	GetRefreshToken(ctx context.Context, token string) (uuid.UUID, error) // Returns user ID if valid
}

// AuthService defines the contract for authentication logic
type AuthService interface {
	HashPassword(password string) (string, error)
	ComparePassword(password, hash string) error
	GenerateTokenPair(user *domain.User) (*domain.TokenPair, error)
	ValidateToken(token string) (uuid.UUID, error) // Returns user ID if valid
	RefreshToken(token string) (*domain.TokenPair, error)
}

// UserService defines the business logic contract for authentication/users
type UserService interface {
	Register(ctx context.Context, email, password, fullName, phone string) (*domain.User, *domain.TokenPair, error)
	Login(ctx context.Context, email, password string) (*domain.TokenPair, error)
	RefreshToken(ctx context.Context, token string) (*domain.TokenPair, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error)
}
