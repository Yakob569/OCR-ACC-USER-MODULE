package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cashflow/auth-service/internal/core/ports"
)

type UserHandler struct {
	svc ports.UserService
}

func NewUserHandler(svc ports.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Only POST is allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	user, err := h.svc.Register(r.Context(), body.Email, body.FullName, body.Phone, body.Role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
