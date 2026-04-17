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

// UserService defines the business logic contract for authentication/users
type UserService interface {
	Register(ctx context.Context, email, fullName, phone, role string) (*domain.User, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error)
}
