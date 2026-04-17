package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	FullName      string    `json:"full_name"`
	Phone         *string   `json:"phone"`
	Role          string    `json:"role"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	AuthProvider  string    `json:"auth_provider"`
	AvatarURL     *string   `json:"avatar_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
