package repositories

import (
	"context"
	"errors"
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
		return nil, err
	}

	if phone.Valid {
		registered.Phone = &phone.String
	}
	if avatar.Valid {
		registered.AvatarURL = &avatar.String
	}

	return &registered, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
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

	return &u, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
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
		return nil, err
	}

	if phone.Valid {
		u.Phone = &phone.String
	}
	if avatar.Valid {
		u.AvatarURL = &avatar.String
	}

	return &u, nil
}

func (r *userRepo) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, userID, token, expiresAt)
	return err
}

func (r *userRepo) RevokeRefreshToken(ctx context.Context, token string) error {
	query := `
		UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1
	`
	_, err := r.db.Exec(ctx, query, token)
	return err
}

func (r *userRepo) GetRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	query := `
		SELECT user_id, expires_at, revoked FROM refresh_tokens WHERE token = $1
	`
	var userID uuid.UUID
	var expiresAt time.Time
	var revoked bool

	err := r.db.QueryRow(ctx, query, token).Scan(&userID, &expiresAt, &revoked)
	if err != nil {
		return uuid.Nil, err
	}

	if revoked {
		return uuid.Nil, errors.New("token revoked")
	}

	if time.Now().After(expiresAt) {
		return uuid.Nil, errors.New("token expired")
	}

	return userID, nil
}
