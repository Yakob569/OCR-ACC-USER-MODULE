package repositories

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new instance of the UserRepository adapter
func NewUserRepository(db *pgxpool.Pool) ports.UserRepository {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) CreateUser(ctx context.Context, u *domain.User) (*domain.User, error) {
	log.Printf("[UserRepo] Executing CreateUser for email: %s", u.Email)
	query := `
		INSERT INTO users (email, full_name, phone, role, auth_provider, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, full_name, phone, role, is_active, email_verified, auth_provider, avatar_url, created_at, updated_at
	`

	var registered domain.User
	var phone, avatar pgtype.Text

	err := r.db.QueryRow(ctx, query,
		u.Email,
		u.FullName,
		u.Phone,
		u.Role,
		u.AuthProvider,
		u.PasswordHash,
	).Scan(
		&registered.ID,
		&registered.Email,
		&registered.FullName,
		&phone,
		&registered.Role,
		&registered.IsActive,
		&registered.EmailVerified,
		&registered.AuthProvider,
		&avatar,
		&registered.CreatedAt,
		&registered.UpdatedAt,
	)

	if err != nil {
		log.Printf("[UserRepo] CreateUser failed: %v", err)
		return nil, err
	}

	if phone.Valid {
		registered.Phone = &phone.String
	}
	if avatar.Valid {
		registered.AvatarURL = &avatar.String
	}

	log.Printf("[UserRepo] CreateUser successful, ID: %s", registered.ID)
	return &registered, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	log.Printf("[UserRepo] Fetching user by email: %s", email)
	query := `
		SELECT id, email, password_hash, full_name, phone, role, is_active, email_verified, auth_provider, avatar_url, created_at, updated_at
		FROM users WHERE email = $1
	`

	var u domain.User
	var phone, avatar, pwdHash pgtype.Text

	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&pwdHash,
		&u.FullName,
		&phone,
		&u.Role,
		&u.IsActive,
		&u.EmailVerified,
		&u.AuthProvider,
		&avatar,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		log.Printf("[UserRepo] GetUserByEmail for %s error/not found: %v", email, err)
		return nil, err
	}

	if pwdHash.Valid {
		u.PasswordHash = pwdHash.String
	}
	if phone.Valid {
		u.Phone = &phone.String
	}
	if avatar.Valid {
		u.AvatarURL = &avatar.String
	}

	log.Printf("[UserRepo] Found user: %s (ID: %s)", email, u.ID)
	return &u, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	log.Printf("[UserRepo] Fetching user by ID: %s", id)
	query := `
		SELECT id, email, full_name, phone, role, is_active, email_verified, auth_provider, avatar_url, created_at, updated_at
		FROM users WHERE id = $1
	`

	var u domain.User
	var phone, avatar pgtype.Text

	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.FullName,
		&phone,
		&u.Role,
		&u.IsActive,
		&u.EmailVerified,
		&u.AuthProvider,
		&avatar,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		log.Printf("[UserRepo] GetUserByID failed for %s: %v", id, err)
		return nil, err
	}

	if phone.Valid {
		u.Phone = &phone.String
	}
	if avatar.Valid {
		u.AvatarURL = &avatar.String
	}

	log.Printf("[UserRepo] Found user: %s for ID: %s", u.Email, id)
	return &u, nil
}

func (r *userRepo) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	log.Printf("[UserRepo] Storing refresh token for user ID: %s", userID)
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		log.Printf("[UserRepo] StoreRefreshToken failed: %v", err)
	}
	return err
}

func (r *userRepo) RevokeRefreshToken(ctx context.Context, token string) error {
	log.Printf("[UserRepo] Revoking refresh token")
	query := `
		UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1
	`
	_, err := r.db.Exec(ctx, query, token)
	if err != nil {
		log.Printf("[UserRepo] RevokeRefreshToken failed: %v", err)
	}
	return err
}

func (r *userRepo) GetRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	log.Printf("[UserRepo] Validating refresh token")
	query := `
		SELECT user_id, expires_at, revoked FROM refresh_tokens WHERE token = $1
	`
	var userID uuid.UUID
	var expiresAt time.Time
	var revoked bool

	err := r.db.QueryRow(ctx, query, token).Scan(&userID, &expiresAt, &revoked)
	if err != nil {
		log.Printf("[UserRepo] GetRefreshToken failed (token might not exist): %v", err)
		return uuid.Nil, err
	}

	if revoked {
		log.Printf("[UserRepo] Refresh token is revoked for user ID: %s", userID)
		return uuid.Nil, errors.New("token revoked")
	}

	if time.Now().After(expiresAt) {
		log.Printf("[UserRepo] Refresh token is expired for user ID: %s", userID)
		return uuid.Nil, errors.New("token expired")
	}

	log.Printf("[UserRepo] Refresh token is valid for user ID: %s", userID)
	return userID, nil
}
