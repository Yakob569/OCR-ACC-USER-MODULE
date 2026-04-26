package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
)

type GroupHandler struct {
	svc       ports.ReceiptGroupService
	uploadSvc ports.ReceiptUploadService
	querySvc  ports.ReceiptQueryService
	reviewSvc ports.ReceiptReviewService
	retrySvc  ports.OCRRetryService
	exportSvc ports.GroupExportService
}

func NewGroupHandler(svc ports.ReceiptGroupService, uploadSvc ports.ReceiptUploadService, querySvc ports.ReceiptQueryService, reviewSvc ports.ReceiptReviewService, retrySvc ports.OCRRetryService, exportSvc ports.GroupExportService) *GroupHandler {
	return &GroupHandler{svc: svc, uploadSvc: uploadSvc, querySvc: querySvc, reviewSvc: reviewSvc, retrySvc: retrySvc, exportSvc: exportSvc}
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

func (h *GroupHandler) UploadGroupImages(w http.ResponseWriter, r *http.Request) {
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

	groupID, err := nestedGroupIDFromPath(r.URL.Path, "images")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid group ID"})
		return
	}

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid multipart form data"})
		return
	}

	fileHeaders := r.MultipartForm.File["files"]
	if len(fileHeaders) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "At least one file is required"})
		return
	}

	files, err := readReceiptFiles(fileHeaders)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	result, err := h.uploadSvc.UploadGroupImages(r.Context(), groupID, userID, files)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Status bool                       `json:"status"`
		Data   *ports.ReceiptUploadResult `json:"data"`
	}{
		Status: true,
		Data:   result,
	})
}

func (h *GroupHandler) ListGroupImages(w http.ResponseWriter, r *http.Request) {
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

	groupID, err := nestedGroupIDFromPath(r.URL.Path, "images")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid group ID"})
		return
	}

	images, err := h.querySvc.ListGroupImages(r.Context(), userID, groupID, queryInt(r, "limit", 20), queryInt(r, "offset", 0))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool                  `json:"status"`
		Data   []domain.ReceiptImage `json:"data"`
	}{Status: true, Data: images})
}

func (h *GroupHandler) ListGroupResults(w http.ResponseWriter, r *http.Request) {
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

	groupID, err := nestedGroupIDFromPath(r.URL.Path, "results")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid group ID"})
		return
	}

	results, err := h.querySvc.ListGroupResults(r.Context(), userID, groupID, queryInt(r, "limit", 20), queryInt(r, "offset", 0))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool                   `json:"status"`
		Data   []domain.OCRExtraction `json:"data"`
	}{Status: true, Data: results})
}

func (h *GroupHandler) GetImage(w http.ResponseWriter, r *http.Request) {
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

	imageID, err := imageIDFromPath(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid image ID"})
		return
	}

	image, err := h.querySvc.GetImage(r.Context(), userID, imageID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Image not found"})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool                 `json:"status"`
		Data   *domain.ReceiptImage `json:"data"`
	}{Status: true, Data: image})
}

func (h *GroupHandler) GetImageResult(w http.ResponseWriter, r *http.Request) {
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

	imageID, err := nestedImageIDFromPath(r.URL.Path, "result")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid image ID"})
		return
	}

	result, err := h.querySvc.GetImageResult(r.Context(), userID, imageID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Result not found"})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool                  `json:"status"`
		Data   *domain.OCRExtraction `json:"data"`
	}{Status: true, Data: result})
}

func (h *GroupHandler) SubmitImageReview(w http.ResponseWriter, r *http.Request) {
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

	imageID, err := nestedImageIDFromPath(r.URL.Path, "review")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid image ID"})
		return
	}

	var body SubmitReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	review, err := h.reviewSvc.SubmitReview(r.Context(), domain.SubmitReceiptReviewInput{
		ReceiptImageID:      imageID,
		ReviewedByUserID:    userID,
		QualityLabel:        body.QualityLabel,
		IsAccepted:          body.IsAccepted,
		CorrectedFieldsJSON: body.CorrectedFields,
		ReviewNotes:         body.ReviewNotes,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Status bool                  `json:"status"`
		Data   *domain.ReceiptReview `json:"data"`
	}{Status: true, Data: review})
}

func (h *GroupHandler) RetryImage(w http.ResponseWriter, r *http.Request) {
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

	imageID, err := nestedImageIDFromPath(r.URL.Path, "retry")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid image ID"})
		return
	}

	job, err := h.retrySvc.RetryImage(r.Context(), userID, imageID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Status bool           `json:"status"`
		Data   *domain.OCRJob `json:"data"`
	}{Status: true, Data: job})
}

func (h *GroupHandler) CreateCSVExport(w http.ResponseWriter, r *http.Request) {
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

	groupID, err := nestedGroupIDFromExportPath(r.URL.Path, "csv")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid group ID"})
		return
	}

	var body CreateCSVExportRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	exportRecord, err := h.exportSvc.CreateCSVExport(r.Context(), userID, groupID, body.SelectedColumns, body.IncludeCorrectedValues)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Status bool                `json:"status"`
		Data   *domain.GroupExport `json:"data"`
	}{Status: true, Data: exportRecord})
}

func (h *GroupHandler) ListGroupExports(w http.ResponseWriter, r *http.Request) {
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

	groupID, err := nestedGroupIDFromPath(r.URL.Path, "exports")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid group ID"})
		return
	}

	exports, err := h.exportSvc.ListGroupExports(r.Context(), userID, groupID, queryInt(r, "limit", 20), queryInt(r, "offset", 0))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool                 `json:"status"`
		Data   []domain.GroupExport `json:"data"`
	}{Status: true, Data: exports})
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

func nestedGroupIDFromPath(path string, tail string) (uuid.UUID, error) {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 5 || parts[4] != tail {
		return uuid.Nil, fmt.Errorf("invalid nested group path")
	}
	return uuid.Parse(parts[3])
}

func readReceiptFiles(fileHeaders []*multipart.FileHeader) ([]ports.ReceiptFile, error) {
	files := make([]ports.ReceiptFile, 0, len(fileHeaders))
	for _, header := range fileHeaders {
		file, err := header.Open()
		if err != nil {
			return nil, err
		}

		bytes, readErr := io.ReadAll(file)
		closeErr := file.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}

		contentType := header.Header.Get("Content-Type")
		files = append(files, ports.ReceiptFile{
			Filename:      header.Filename,
			ContentType:   contentType,
			ContentLength: header.Size,
			Bytes:         bytes,
		})
	}

	return files, nil
}

func imageIDFromPath(path string) (uuid.UUID, error) {
	idPart := strings.TrimPrefix(path, "/api/v1/images/")
	return uuid.Parse(strings.TrimSpace(idPart))
}

func nestedImageIDFromPath(path string, tail string) (uuid.UUID, error) {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 5 || parts[4] != tail {
		return uuid.Nil, fmt.Errorf("invalid nested image path")
	}
	return uuid.Parse(parts[3])
}

func nestedGroupIDFromExportPath(path string, tail string) (uuid.UUID, error) {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 6 || parts[4] != "exports" || parts[5] != tail {
		return uuid.Nil, fmt.Errorf("invalid export path")
	}
	return uuid.Parse(parts[3])
}
