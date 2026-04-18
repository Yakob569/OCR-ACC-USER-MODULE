package ports

import (
	"context"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/google/uuid"
)

// UserRepository defines the persistence contract for users
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

// AuthService defines the contract for external authentication providers (e.g. Supabase)
type AuthService interface {
	SignUp(ctx context.Context, email, password string) (string, error) // Returns external ID
	Login(ctx context.Context, email, password string) (string, error)  // Returns token
	SignOut(ctx context.Context) error
	ResetPassword(ctx context.Context, email string) error
	GetSocialLoginURL(provider string) (string, error)
}

// UserService defines the business logic contract for authentication/users
type UserService interface {
	Register(ctx context.Context, email, password, fullName, phone string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ResetPassword(ctx context.Context, email string) error
	GetSocialLoginURL(provider string) (string, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error)
}
