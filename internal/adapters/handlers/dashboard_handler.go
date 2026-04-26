package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cashflow/auth-service/internal/core/ports"
)

type DashboardHandler struct {
	svc ports.DashboardService
}

func NewDashboardHandler(svc ports.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Only GET is allowed"})
		return
	}

	userID, ok := requestUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Unauthorized"})
		return
	}

	summary, err := h.svc.GetSummary(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Failed to load dashboard summary"})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool        `json:"status"`
		Data   interface{} `json:"data"`
	}{
		Status: true,
		Data:   summary,
	})
}
