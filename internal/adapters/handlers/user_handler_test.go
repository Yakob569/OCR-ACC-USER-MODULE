package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/google/uuid"
)

type mockUserService struct {
	refreshTokenFn func(ctx context.Context, token string) (*domain.TokenPair, error)
	logoutFn       func(ctx context.Context, token string) error
}

func (m mockUserService) Register(ctx context.Context, email, password, fullName, phone string) (*domain.User, *domain.TokenPair, error) {
	panic("unexpected call to Register")
}
func (m mockUserService) Login(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	panic("unexpected call to Login")
}
func (m mockUserService) RefreshToken(ctx context.Context, token string) (*domain.TokenPair, error) {
	if m.refreshTokenFn == nil {
		panic("unexpected call to RefreshToken")
	}
	return m.refreshTokenFn(ctx, token)
}
func (m mockUserService) Logout(ctx context.Context, token string) error {
	if m.logoutFn == nil {
		panic("unexpected call to Logout")
	}
	return m.logoutFn(ctx, token)
}
func (m mockUserService) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	panic("unexpected call to GetProfile")
}

func TestUserHandler_RefreshToken_EmptyToken_Returns400(t *testing.T) {
	h := NewUserHandler(mockUserService{
		refreshTokenFn: func(ctx context.Context, token string) (*domain.TokenPair, error) {
			t.Fatalf("RefreshToken should not be called for empty token")
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/refresh", strings.NewReader(`{"refresh_token":""}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.RefreshToken(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestUserHandler_RefreshToken_ShortToken_NoPanic(t *testing.T) {
	h := NewUserHandler(mockUserService{
		refreshTokenFn: func(ctx context.Context, token string) (*domain.TokenPair, error) {
			return nil, errors.New("invalid token")
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/refresh", strings.NewReader(`{"refresh_token":"abc"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.RefreshToken(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestUserHandler_Logout_EmptyToken_Returns400(t *testing.T) {
	h := NewUserHandler(mockUserService{
		logoutFn: func(ctx context.Context, token string) error {
			t.Fatalf("Logout should not be called for empty token")
			return nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/logout", strings.NewReader(`{"refresh_token":"   "}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestTokenPrefixAndLen(t *testing.T) {
	prefix, ln := tokenPrefixAndLen("  abcdef  ", 3)
	if prefix != "abc" || ln != 6 {
		t.Fatalf("unexpected result: prefix=%q len=%d", prefix, ln)
	}

	prefix, ln = tokenPrefixAndLen("", 10)
	if prefix != "" || ln != 0 {
		t.Fatalf("unexpected empty result: prefix=%q len=%d", prefix, ln)
	}

	prefix, ln = tokenPrefixAndLen("ab", 10)
	if prefix != "ab" || ln != 2 {
		t.Fatalf("unexpected short result: prefix=%q len=%d", prefix, ln)
	}

	prefix, ln = tokenPrefixAndLen("abcd", 0)
	if prefix != "" || ln != 4 {
		t.Fatalf("unexpected maxPrefix=0 result: prefix=%q len=%d", prefix, ln)
	}

	// ensure deterministic on whitespace-only
	prefix, ln = tokenPrefixAndLen("   ", 10)
	if prefix != "" || ln != 0 {
		t.Fatalf("unexpected whitespace result: prefix=%q len=%d", prefix, ln)
	}
}
