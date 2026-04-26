package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
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

	if strings.Contains(errStr, "invalid email") {
		return "The email address provided is invalid."
	}
	if strings.Contains(errStr, "already exists") {
		return "This email is already registered."
	}
	if strings.Contains(errStr, "password is too short") {
		return "Password must be at least 6 characters long."
	}

	// Default generic message for safety
	return "Internal server error. Please try again later."
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("➡️  [POST] /api/v1/register - New registration attempt")

	if r.Method != http.MethodPost {
		log.Printf("⚠️  [Register] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Only POST is allowed"})
		return
	}

	var body RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("❌ [Register] JSON decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	log.Printf("ℹ️  [Register] Attempting to register email: %s", body.Email)
	user, tokens, err := h.svc.Register(r.Context(), body.Email, body.Password, body.FullName, body.Phone)
	if err != nil {
		log.Printf("❌ [Register] Service error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: h.mapError(err)})
		return
	}

	log.Printf("✅ [Register] Successfully registered user ID: %s", user.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Status bool        `json:"status"`
		Data   interface{} `json:"data"`
		Tokens interface{} `json:"tokens"`
	}{
		Status: true,
		Data:   user,
		Tokens: tokens,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("➡️  [POST] /api/v1/login - Login attempt")

	if r.Method != http.MethodPost {
		log.Printf("⚠️  [Login] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Only POST is allowed"})
		return
	}

	var body LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("❌ [Login] JSON decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	log.Printf("ℹ️  [Login] Authenticating email: %s", body.Email)
	tokens, err := h.svc.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		log.Printf("❌ [Login] Authentication failed: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid email or password"})
		return
	}

	log.Printf("✅ [Login] Successful login for email: %s", body.Email)
	json.NewEncoder(w).Encode(TokenResponse{
		Status:       true,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	})
}

func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("➡️  [POST] /api/v1/refresh - Token refresh request")

	if r.Method != http.MethodPost {
		log.Printf("⚠️  [Refresh] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Only POST is allowed"})
		return
	}

	var body RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("❌ [Refresh] JSON decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	tokens, err := h.svc.RefreshToken(r.Context(), body.RefreshToken)
	if err != nil {
		log.Printf("❌ [Refresh] Refresh failed: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid or expired refresh token"})
		return
	}

	log.Printf("✅ [Refresh] Token rotated successfully")
	json.NewEncoder(w).Encode(TokenResponse{
		Status:       true,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("➡️  [POST] /api/v1/logout - Logout request")

	if r.Method != http.MethodPost {
		log.Printf("⚠️  [Logout] Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Only POST is allowed"})
		return
	}

	var body LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("❌ [Logout] JSON decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	if err := h.svc.Logout(r.Context(), body.RefreshToken); err != nil {
		log.Printf("❌ [Logout] Logout failed: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid or expired refresh token"})
		return
	}

	log.Printf("✅ [Logout] Refresh token revoked successfully")
	json.NewEncoder(w).Encode(MessageResponse{
		Status:  true,
		Message: "Logged out successfully",
	})
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("➡️  [GET] /api/v1/profile - Fetching user profile")

	val := r.Context().Value("user_id")
	if val == nil {
		log.Printf("⚠️  [Profile] Unauthorized access attempt (no user_id in context)")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Unauthorized"})
		return
	}

	userID, ok := val.(uuid.UUID)
	if !ok {
		log.Printf("❌ [Profile] Invalid context user_id type")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid user ID in context"})
		return
	}

	user, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		log.Printf("❌ [Profile] User not found for ID: %s", userID)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "User not found"})
		return
	}

	log.Printf("✅ [Profile] Profile retrieved for user: %s", user.Email)
	json.NewEncoder(w).Encode(struct {
		Status bool        `json:"status"`
		Data   interface{} `json:"data"`
	}{
		Status: true,
		Data:   user,
	})
}
