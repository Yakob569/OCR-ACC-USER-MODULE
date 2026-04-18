package auth

import (
	"context"
	"fmt"

	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/supabase-community/gotrue-go"
	"github.com/supabase-community/gotrue-go/types"
)

type supabaseAuthAdapter struct {
	client gotrue.Client
	url    string
}

func NewSupabaseAuthAdapter(projectRef, url, key string) ports.AuthService {
	client := gotrue.New(projectRef, key)
	return &supabaseAuthAdapter{
		client: client,
		url:    url,
	}
}

func (a *supabaseAuthAdapter) GetSocialLoginURL(provider string) (string, error) {
	// Constructing the Supabase OAuth URL
	// Format: https://<project-url>/auth/v1/authorize?provider=<provider>
	socialURL := fmt.Sprintf("%s/auth/v1/authorize?provider=%s", a.url, provider)
	return socialURL, nil
}

func (a *supabaseAuthAdapter) SignUp(ctx context.Context, email, password string) (string, error) {
	resp, err := a.client.Signup(types.SignupRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return resp.User.ID.String(), nil
}

func (a *supabaseAuthAdapter) Login(ctx context.Context, email, password string) (string, error) {
	resp, err := a.client.SignInWithEmailPassword(email, password)
	if err != nil {
		return "", err
	}
	return resp.AccessToken, nil
}

func (a *supabaseAuthAdapter) SignOut(ctx context.Context) error {
	return a.client.Logout()
}

func (a *supabaseAuthAdapter) ResetPassword(ctx context.Context, email string) error {
	return a.client.Recover(types.RecoverRequest{
		Email: email,
	})
}
