package handlers

import "encoding/json"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type RegisterResponse struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"` // Using interface{} to allow domain.User without direct import loop if possible, or just use any
}

type TokenResponse struct {
	Status       bool   `json:"status"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type SocialLoginResponse struct {
	Status           bool   `json:"status"`
	AuthorizationURL string `json:"authorization_url"`
}

type ErrorResponse struct {
	Status bool   `json:"status"`
	Error  string `json:"error"`
}

type MessageResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type CreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SubmitReviewRequest struct {
	QualityLabel    string          `json:"quality_label"`
	IsAccepted      bool            `json:"is_accepted"`
	CorrectedFields json.RawMessage `json:"corrected_fields"`
	ReviewNotes     string          `json:"review_notes"`
}
