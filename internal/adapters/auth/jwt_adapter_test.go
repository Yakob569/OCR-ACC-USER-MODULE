package auth

import (
	"testing"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/google/uuid"
)

func TestJWTAuthAdapter(t *testing.T) {
	secret := "test-secret"
	adapter := NewJWTAuthAdapter(secret)

	// Test Password Hashing
	password := "secure-password"
	hash, err := adapter.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	err = adapter.ComparePassword(password, hash)
	if err != nil {
		t.Errorf("Password comparison failed: %v", err)
	}

	err = adapter.ComparePassword("wrong-password", hash)
	if err == nil {
		t.Errorf("Password comparison should have failed for wrong password")
	}

	// Test Token Generation
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "user",
	}

	tokens, err := adapter.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Errorf("Access token is empty")
	}

	if tokens.RefreshToken == "" {
		t.Errorf("Refresh token is empty")
	}

	// Test Token Validation
	userID, err := adapter.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if userID != user.ID {
		t.Errorf("Validated user ID mismatch: expected %v, got %v", user.ID, userID)
	}

	_, err = adapter.ValidateToken(tokens.RefreshToken)
	if err == nil {
		t.Errorf("Refresh token should not be accepted as an access token")
	}
}
