package auth

import (
	"errors"
	"time"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type jwtAuthAdapter struct {
	secretKey     []byte
	accessTokenExp time.Duration
}

func NewJWTAuthAdapter(secret string) ports.AuthService {
	return &jwtAuthAdapter{
		secretKey:      []byte(secret),
		accessTokenExp: time.Hour * 1, // 1 hour access token
	}
}

func (a *jwtAuthAdapter) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (a *jwtAuthAdapter) ComparePassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (a *jwtAuthAdapter) GenerateTokenPair(user *domain.User) (*domain.TokenPair, error) {
	// Access Token
	accessClaims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(a.accessTokenExp).Unix(),
		"iat":   time.Now().Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(a.secretKey)
	if err != nil {
		return nil, err
	}

	// Refresh Token (simplified for this implementation, just a random UUID or longer JWT)
	refreshClaims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days refresh token
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(a.secretKey)
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresAt:    accessClaims["exp"].(int64),
	}, nil
}

func (a *jwtAuthAdapter) ValidateToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return a.secretKey, nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, errors.New("sub claim missing")
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func (a *jwtAuthAdapter) RefreshToken(tokenStr string) (*domain.TokenPair, error) {
	// Logic to refresh token by validating the refresh token and generating a new pair
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return a.secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	return nil, errors.New("use GenerateTokenPair instead")
}
