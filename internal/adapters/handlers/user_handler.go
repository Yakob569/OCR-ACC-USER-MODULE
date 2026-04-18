package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/cashflow/auth-service/internal/core/ports"
)

type UserHandler struct {
	svc ports.UserService
}

func NewUserHandler(svc ports.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// mapError converts internal errors to user-friendly ones if they are safe to disclose
func (h *UserHandler) mapError(err error) string {
	errStr := err.Error()

	// Check for Supabase specific validation errors
	if strings.Contains(errStr, "email_address_invalid") || strings.Contains(errStr, "invalid email") {
		return "The email address provided is invalid."
	}
	if strings.Contains(errStr, "already registered") || strings.Contains(errStr, "already exists") {
		return "This email is already registered."
	}
	if strings.Contains(errStr, "password is too short") {
		return "Password must be at least 6 characters long."
	}
	if strings.Contains(errStr, "over_email_send_rate_limit") || strings.Contains(errStr, "rate limit exceeded") {
		return "Too many requests. Please try again in an hour."
	}

	// Default generic message for safety
	return "Internal server error. Please try again later."
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Only POST is allowed"})
		return
	}

	var body RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	user, err := h.svc.Register(r.Context(), body.Email, body.Password, body.FullName, body.Phone)
	if err != nil {
		log.Printf("[Handler] Register Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: h.mapError(err)})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{Status: true, Data: user})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Only POST is allowed"})
		return
	}

	var body LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	token, err := h.svc.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		log.Printf("[Handler] Login Error: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid email or password"})
		return
	}

	json.NewEncoder(w).Encode(TokenResponse{Status: true, AccessToken: token})
}

func (h *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Only POST is allowed"})
		return
	}

	var body ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	err := h.svc.ResetPassword(r.Context(), body.Email)
	if err != nil {
		log.Printf("[Handler] ForgotPassword Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: h.mapError(err)})
		return
	}

	json.NewEncoder(w).Encode(MessageResponse{Status: true, Message: "Password reset instructions sent to email"})
}

func (h *UserHandler) SocialLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "provider query parameter is required"})
		return
	}

	url, err := h.svc.GetSocialLoginURL(provider)
	if err != nil {
		log.Printf("[Handler] SocialLogin Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Failed to generate authorization URL"})
		return
	}

	json.NewEncoder(w).Encode(SocialLoginResponse{Status: true, AuthorizationURL: url})
}
