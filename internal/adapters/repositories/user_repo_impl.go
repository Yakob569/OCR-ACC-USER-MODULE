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
	log.Printf("➡️  [UserRepo.CreateUser] Attempting to create user with email: %s", u.Email)
	if r.db == nil {
		log.Printf("❌ [UserRepo.CreateUser] Database connection is not available for email: %s", u.Email)
		return nil, errors.New("database connection is not available")
	}

	query := `
		INSERT INTO users (email, full_name, phone, role, auth_provider, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, full_name, phone, role, is_active, email_verified, auth_provider, avatar_url, created_at, updated_at
	`

	var registered domain.User
	var phone, avatar pgtype.Text

	log.Printf("ℹ️  [UserRepo.CreateUser] Executing SQL query for email: %s", u.Email)
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
		log.Printf("❌ [UserRepo.CreateUser] Failed to create user %s: %v", u.Email, err)
		return nil, err
	}

	if phone.Valid {
		registered.Phone = &phone.String
	}
	if avatar.Valid {
		registered.AvatarURL = &avatar.String
	}

	log.Printf("✅ [UserRepo.CreateUser] User created successfully with ID: %s, Email: %s", registered.ID, registered.Email)
	return &registered, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	log.Printf("➡️  [UserRepo.GetUserByEmail] Attempting to fetch user by email: %s", email)
	if r.db == nil {
		log.Printf("❌ [UserRepo.GetUserByEmail] Database connection is not available for email: %s", email)
		return nil, errors.New("database connection is not available")
	}
	query := `
		SELECT id, email, password_hash, full_name, phone, role, is_active, email_verified, auth_provider, avatar_url, created_at, updated_at
		FROM users WHERE email = $1
	`

	var u domain.User
	var phone, avatar, pwdHash pgtype.Text

	log.Printf("ℹ️  [UserRepo.GetUserByEmail] Executing SQL query for email: %s", email)
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
		log.Printf("❌ [UserRepo.GetUserByEmail] Failed to fetch user by email %s: %v", email, err)
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

	log.Printf("✅ [UserRepo.GetUserByEmail] User found with ID: %s, Email: %s", u.ID, u.Email)
	return &u, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	log.Printf("➡️  [UserRepo.GetUserByID] Attempting to fetch user by ID: %s", id)
	if r.db == nil {
		log.Printf("❌ [UserRepo.GetUserByID] Database connection is not available for ID: %s", id)
		return nil, errors.New("database connection is not available")
	}
	query := `
		SELECT id, email, full_name, phone, role, is_active, email_verified, auth_provider, avatar_url, created_at, updated_at
		FROM users WHERE id = $1
	`

	var u domain.User
	var phone, avatar pgtype.Text

	log.Printf("ℹ️  [UserRepo.GetUserByID] Executing SQL query for user ID: %s", id)
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
		log.Printf("❌ [UserRepo.GetUserByID] Failed to fetch user for ID %s: %v", id, err)
		return nil, err
	}

	if phone.Valid {
		u.Phone = &phone.String
	}
	if avatar.Valid {
		u.AvatarURL = &avatar.String
	}

	log.Printf("✅ [UserRepo.GetUserByID] User found with ID: %s, Email: %s", u.ID, u.Email)
	return &u, nil
}

func (r *userRepo) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	log.Printf("➡️  [UserRepo.StoreRefreshToken] Attempting to store refresh token for user ID: %s", userID)
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`
	log.Printf("ℹ️  [UserRepo.StoreRefreshToken] Executing SQL query for user ID: %s", userID)
	_, err := r.db.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		log.Printf("❌ [UserRepo.StoreRefreshToken] Failed to store refresh token for user ID %s: %v", userID, err)
		return err
	}
	log.Printf("✅ [UserRepo.StoreRefreshToken] Refresh token stored successfully for user ID: %s", userID)
	return nil
}

func (r *userRepo) RevokeRefreshToken(ctx context.Context, token string) error {
	log.Printf("➡️  [UserRepo.RevokeRefreshToken] Attempting to revoke refresh token (truncated): %s...", token[:10])
	query := `
		UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1
	`
	log.Printf("ℹ️  [UserRepo.RevokeRefreshToken] Executing SQL query to revoke token")
	_, err := r.db.Exec(ctx, query, token)
	if err != nil {
		log.Printf("❌ [UserRepo.RevokeRefreshToken] Failed to revoke refresh token: %v", err)
		return err
	}
	log.Printf("✅ [UserRepo.RevokeRefreshToken] Refresh token revoked successfully")
	return nil
}

func (r *userRepo) GetRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	log.Printf("➡️  [UserRepo.GetRefreshToken] Attempting to validate refresh token (truncated): %s...", token[:10])
	query := `
		SELECT user_id, expires_at, revoked FROM refresh_tokens WHERE token = $1
	`
	var userID uuid.UUID
	var expiresAt time.Time
	var revoked bool

	log.Printf("ℹ️  [UserRepo.GetRefreshToken] Executing SQL query to fetch refresh token details")
	err := r.db.QueryRow(ctx, query, token).Scan(&userID, &expiresAt, &revoked)
	if err != nil {
		log.Printf("❌ [UserRepo.GetRefreshToken] Failed to fetch refresh token details: %v", err)
		return uuid.Nil, err
	}

	if revoked {
		log.Printf("❌ [UserRepo.GetRefreshToken] Refresh token is revoked for user ID: %s", userID)
		return uuid.Nil, errors.New("token revoked")
	}

	if time.Now().After(expiresAt) {
		log.Printf("❌ [UserRepo.GetRefreshToken] Refresh token is expired for user ID: %s", userID)
		return uuid.Nil, errors.New("token expired")
	}

	log.Printf("✅ [UserRepo.GetRefreshToken] Refresh token is valid for user ID: %s", userID)
	return userID, nil
}
