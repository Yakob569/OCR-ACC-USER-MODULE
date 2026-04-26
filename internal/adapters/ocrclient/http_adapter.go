package ocrclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/cashflow/auth-service/internal/config"
	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
)

type httpOCREngineService struct {
	baseURL string
	client  *http.Client
}

type ocrEngineResponse struct {
	Success     bool            `json:"success"`
	ReceiptType string          `json:"receipt_type"`
	Fields      json.RawMessage `json:"fields"`
	Items       json.RawMessage `json:"items"`
	Warnings    json.RawMessage `json:"warnings"`
	RawText     *string         `json:"raw_text"`
	Debug       json.RawMessage `json:"debug"`
}

func NewOCREngineService(cfg config.OCREngineConfig) ports.OCREngineService {
	return &httpOCREngineService{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		client: &http.Client{
			Timeout: timeDurationSeconds(cfg.TimeoutSeconds),
		},
	}
}

func (s *httpOCREngineService) Extract(ctx context.Context, filename, contentType string, content io.Reader) (*domain.OCRProcessResult, error) {
	if s.baseURL == "" {
		return nil, fmt.Errorf("OCR_ENGINE_BASE_URL is not configured")
	}

	fileBytes, err := io.ReadAll(content)
	if err != nil {
		return nil, err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(fileBytes); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/api/v1/ocr/extract", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if contentType != "" {
		req.Header.Set("X-Original-Content-Type", contentType)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OCR engine returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var payload ocrEngineResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	receiptType := stringPtrIfNonEmpty(payload.ReceiptType)
	extraction := &domain.OCRExtraction{
		Success:      payload.Success,
		ReceiptType:  receiptType,
		FieldsJSON:   normalizeJSON(payload.Fields, []byte(`{}`)),
		ItemsJSON:    normalizeJSON(payload.Items, []byte(`[]`)),
		WarningsJSON: normalizeJSON(payload.Warnings, []byte(`[]`)),
		RawText:      payload.RawText,
		DebugJSON:    normalizeJSON(payload.Debug, nil),
	}

	status := domain.OCRStatusCompleted
	if !payload.Success {
		status = domain.OCRStatusNeedsReview
	}

	confidence := deriveOverallConfidence(payload.Fields, payload.Items)
	return &domain.OCRProcessResult{
		Extraction:  extraction,
		OCRStatus:   status,
		ReceiptType: receiptType,
		Confidence:  confidence,
	}, nil
}

func normalizeJSON(raw json.RawMessage, fallback []byte) []byte {
	if len(raw) == 0 {
		return fallback
	}
	return raw
}

func stringPtrIfNonEmpty(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func deriveOverallConfidence(fieldsRaw, itemsRaw json.RawMessage) *float64 {
	type fieldValue struct {
		Confidence float64 `json:"confidence"`
	}
	type itemValue struct {
		Confidence float64 `json:"confidence"`
	}

	var total float64
	var count int

	var fields map[string]fieldValue
	if len(fieldsRaw) > 0 && json.Unmarshal(fieldsRaw, &fields) == nil {
		for _, field := range fields {
			total += field.Confidence
			count++
		}
	}

	var items []itemValue
	if len(itemsRaw) > 0 && json.Unmarshal(itemsRaw, &items) == nil {
		for _, item := range items {
			total += item.Confidence
			count++
		}
	}

	if count == 0 {
		return nil
	}

	avg := total / float64(count)
	return &avg
}

func timeDurationSeconds(seconds int) time.Duration {
	if seconds <= 0 {
		seconds = 60
	}
	return time.Duration(seconds) * time.Second
}
