package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
)

type GroupHandler struct {
	svc ports.ReceiptGroupService
}

func NewGroupHandler(svc ports.ReceiptGroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Only POST is allowed"})
		return
	}

	userID, ok := requestUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Unauthorized"})
		return
	}

	var body CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	group, err := h.svc.CreateGroup(r.Context(), domain.CreateReceiptGroupInput{
		UserID:      userID,
		Name:        body.Name,
		Description: body.Description,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Status bool                 `json:"status"`
		Data   *domain.ReceiptGroup `json:"data"`
	}{
		Status: true,
		Data:   group,
	})
}

func (h *GroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
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

	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	groups, err := h.svc.ListGroups(r.Context(), userID, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Failed to list groups"})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool                  `json:"status"`
		Data   []domain.ReceiptGroup `json:"data"`
	}{
		Status: true,
		Data:   groups,
	})
}

func (h *GroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
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

	groupID, err := groupIDFromPath(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid group ID"})
		return
	}

	group, err := h.svc.GetGroup(r.Context(), userID, groupID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Group not found"})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool                 `json:"status"`
		Data   *domain.ReceiptGroup `json:"data"`
	}{
		Status: true,
		Data:   group,
	})
}

func requestUserID(r *http.Request) (uuid.UUID, bool) {
	val := r.Context().Value("user_id")
	if val == nil {
		return uuid.Nil, false
	}

	userID, ok := val.(uuid.UUID)
	return userID, ok
}

func queryInt(r *http.Request, key string, fallback int) int {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func groupIDFromPath(path string) (uuid.UUID, error) {
	idPart := strings.TrimPrefix(path, "/api/v1/groups/")
	return uuid.Parse(strings.TrimSpace(idPart))
}
