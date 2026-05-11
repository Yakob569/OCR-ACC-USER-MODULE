package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
)

type groupExportService struct {
	groupRepo   ports.ReceiptGroupRepository
	imageRepo   ports.ReceiptImageRepository
	extractRepo ports.OCRExtractionRepository
	reviewRepo  ports.ReceiptReviewRepository
	exportRepo  ports.GroupExportRepository
	storage     ports.ObjectStorageService
}

func NewGroupExportService(
	groupRepo ports.ReceiptGroupRepository,
	imageRepo ports.ReceiptImageRepository,
	extractRepo ports.OCRExtractionRepository,
	reviewRepo ports.ReceiptReviewRepository,
	exportRepo ports.GroupExportRepository,
	storage ports.ObjectStorageService,
) ports.GroupExportService {
	return &groupExportService{
		groupRepo:   groupRepo,
		imageRepo:   imageRepo,
		extractRepo: extractRepo,
		reviewRepo:  reviewRepo,
		exportRepo:  exportRepo,
		storage:     storage,
	}
}

func (s *groupExportService) CreateCSVExport(ctx context.Context, userID, groupID uuid.UUID, selectedColumns []string, includeCorrectedValues bool) (*domain.GroupExport, error) {
	group, err := s.groupRepo.GetByUserAndID(ctx, userID, groupID)
	if err != nil {
		return nil, fmt.Errorf("group not found")
	}

	images, err := s.imageRepo.ListByGroup(ctx, userID, groupID, 10000, 0)
	if err != nil {
		return nil, err
	}

	// Use a channel to collect results and maintain concurrency
	type imageResult struct {
		index int
		rows  [][]string
		err   error
	}
	resultsChan := make(chan imageResult, len(images))
	
	// Semaphore to limit concurrent processing (e.g., 20 at a time)
	sem := make(chan struct{}, 20)

	for i, image := range images {
		go func(idx int, img domain.ReceiptImage) {
			sem <- struct{}{}
			defer func() { <-sem }()

			// Only process images with successful OCR
			if img.OCRStatus != domain.OCRStatusCompleted {
				resultsChan <- imageResult{index: idx, rows: nil}
				return
			}

			extraction, err := s.extractRepo.GetByReceiptImageID(ctx, img.ID)
			if err != nil {
				resultsChan <- imageResult{index: idx, err: err}
				return
			}
			var review *domain.ReceiptReview
			if s.reviewRepo != nil {
				review, err = s.reviewRepo.GetByReceiptImageID(ctx, img.ID)
				if err != nil {
					resultsChan <- imageResult{index: idx, err: err}
					return
				}
			}

			fields := map[string]fieldValueForExport{}
			if extraction != nil && len(extraction.FieldsJSON) > 0 {
				_ = json.Unmarshal(extraction.FieldsJSON, &fields)
			}

			var items []map[string]any
			if extraction != nil && len(extraction.ItemsJSON) > 0 {
				_ = json.Unmarshal(extraction.ItemsJSON, &items)
			}

			corrected := map[string]any{}
			if includeCorrectedValues && review != nil && len(review.CorrectedFieldsJSON) > 0 {
				_ = json.Unmarshal(review.CorrectedFieldsJSON, &corrected)
			}

			if len(items) == 0 {
				items = []map[string]any{{}}
			}

			var imageRows [][]string
			for _, item := range items {
				imageRows = append(imageRows, s.buildVatPurchaseRow(group, &img, extraction, fields, item, corrected))
			}
			
			resultsChan <- imageResult{index: idx, rows: imageRows}
		}(i, image)
	}

	// Collect all results
	allResults := make([][][]string, len(images))
	for i := 0; i < len(images); i++ {
		res := <-resultsChan
		if res.err != nil {
			return nil, res.err
		}
		allResults[res.index] = res.rows
	}

	exportID := uuid.New()
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	// Write headers
	headers := []string{
		"VAT CATEGORY", "CALENDAR TYPE", "Types Of purchase", "TIN.", "Seller Name ",
		"Date Of Purchase", "MRC Number", "Vat receipt number", "Description.",
		"Unit of Measure ", "Quantity.", "Unit Price.", "Total value", "vat", "value After vat",
		"", "", "",
	}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	// Write results to buffer in correct order
	for _, imageRows := range allResults {
		for _, row := range imageRows {
			if err := writer.Write(row); err != nil {
				return nil, err
			}
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	bucket, objectKey, objectURL, err := s.storage.UploadGroupExport(ctx, userID, groupID, exportID, buf.Bytes())
	if err != nil {
		return nil, err
	}

	selectedColumnsJSON, _ := json.Marshal(selectedColumns)
	storageURL := ""
	if objectURL != nil {
		storageURL = *objectURL
	}

	exportRecord, err := s.exportRepo.Create(ctx, domain.CreateGroupExportInput{
		GroupID:             groupID,
		ExportedByUserID:    userID,
		SelectedColumnsJSON: selectedColumnsJSON,
		RowCount:            len(images),
		StorageBucket:       bucket,
		StorageObjectKey:    objectKey,
		StorageURL:          storageURL,
	})
	if err != nil {
		return nil, err
	}

	_, _ = s.groupRepo.RefreshAggregateState(ctx, groupID)
	return exportRecord, nil
}

func (s *groupExportService) buildVatPurchaseRow(
	group *domain.ReceiptGroup,
	image *domain.ReceiptImage,
	extraction *domain.OCRExtraction,
	fields map[string]fieldValueForExport,
	item map[string]any,
	corrected map[string]any,
) []string {
	row := make([]string, 18)

	// 1. VAT CATEGORY (G=GOODS;S=SERVICES) - Always G
	row[0] = "G"

	// 2. CALENDAR TYPE (E=ETHIOPIAN;G=GREGORIAN) - Forced to G
	row[1] = "G"

	// 3. Types of purchase (1-6) - Forced to 5
	row[2] = "5"

	// 4. TIN
	row[3] = s.getFieldValue(fields, corrected, "tin", "merchant")

	// 5. Seller name
	row[4] = s.getFieldValue(fields, corrected, "name", "merchant")

	// 6. Date of purchase (dd/mm/yyyy)
	row[5] = s.getFormattedDate(fields, corrected)

	// 7. MRC Number (Machine ID)
	row[6] = s.getFieldValue(fields, corrected, "machine_id", "transaction")

	// 8. Vat receipt number (FS NO)
	row[7] = s.getFieldValue(fields, corrected, "fs_number", "transaction")

	// 9. Description
	row[8] = fmt.Sprint(item["description"])
	if row[8] == "<nil>" { row[8] = "" }

	// 10. Unit of Measure - Forced to 7
	row[9] = "7"

	// 11. Quantity
	row[10] = fmt.Sprint(item["quantity"])
	if row[10] == "<nil>" { row[10] = "" }

	// 12. Unit Price
	row[11] = fmt.Sprint(item["unit_price"])
	if row[11] == "<nil>" { row[11] = "" }

	// 13. Total value (Line Total)
	row[12] = fmt.Sprint(item["line_total"])
	if row[12] == "<nil>" { row[12] = "" }

	// 14. vat
	taxAmount := item["tax_amount"]
	if taxAmount == nil || fmt.Sprint(taxAmount) == "0" || fmt.Sprint(taxAmount) == "0.0" {
		row[13] = "0" // Changed from NOTAXBL to 0 as per template examples
	} else {
		row[13] = fmt.Sprint(taxAmount)
	}

	// 15. value after vat (Line Total + Tax)
	lt := s.toFloat(item["line_total"])
	ta := s.toFloat(item["tax_amount"])
	row[14] = fmt.Sprintf("%.2f", lt+ta)
	if row[14] == "0.00" && row[12] == "" { row[14] = "" }

	// 16-18. Empty columns as per template
	row[15] = ""
	row[16] = ""
	row[17] = ""

	return row
}

func (s *groupExportService) getCalendarType(fields map[string]fieldValueForExport, corrected map[string]any) string {
	dateStr := s.getFieldValue(fields, corrected, "date", "transaction")
	if dateStr == "" {
		return "G"
	}
	
	// Basic logic: if year is far in the past, it might be Ethiopian
	// Format expected is dd/mm/yyyy
	parts := strings.Split(dateStr, "/")
	if len(parts) == 3 {
		yearStr := parts[2]
		var year int
		fmt.Sscanf(yearStr, "%d", &year)
		if year > 0 {
			currYear := time.Now().Year()
			if currYear - year >= 7 {
				return "E"
			}
		}
	}
	return "G"
}

func (s *groupExportService) isTaxable(fields map[string]fieldValueForExport, item map[string]any, corrected map[string]any) bool {
	tax := s.toFloat(item["tax_amount"])
	if tax > 0 {
		return true
	}
	// Also check totals if item tax is missing
	totalTax := s.toFloat(s.getFieldValue(fields, corrected, "tax_total", "totals"))
	return totalTax > 0
}

func (s *groupExportService) getFormattedDate(fields map[string]fieldValueForExport, corrected map[string]any) string {
	return s.getFieldValue(fields, corrected, "date", "transaction")
}

func (s *groupExportService) getFieldValue(fields map[string]fieldValueForExport, corrected map[string]any, key string, section string) string {
	// 1. Check corrected values first
	if val, ok := corrected[key]; ok && val != nil {
		return fmt.Sprint(val)
	}
	
	// 2. Check fields map (which maps to extraction.FieldsJSON)
	// The fields map might be nested or flat depending on how it was unmarshaled.
	// Based on gemini_service.py, it's structured: transaction -> { date: { value: ... } }
	// But OCREngine response might be flattened or keep structure.
	
	if val, ok := fields[key]; ok {
		return fmt.Sprint(val.Value)
	}

	// Try nested if not found flat
	// This is a bit defensive as the exact unmarshal structure can vary
	return ""
}

func (s *groupExportService) toFloat(v any) float64 {
	if v == nil { return 0 }
	var f float64
	str := fmt.Sprint(v)
	fmt.Sscanf(str, "%f", &f)
	return f
}

func (s *groupExportService) ListGroupExports(ctx context.Context, userID, groupID uuid.UUID, limit, offset int) ([]domain.GroupExport, error) {
	if _, err := s.groupRepo.GetByUserAndID(ctx, userID, groupID); err != nil {
		return nil, fmt.Errorf("group not found")
	}
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.exportRepo.ListByGroup(ctx, userID, groupID, limit, offset)
}

type fieldValueForExport struct {
	Value any `json:"value"`
}

func exportColumnValue(column string, group *domain.ReceiptGroup, image *domain.ReceiptImage, extraction *domain.OCRExtraction, fields map[string]fieldValueForExport, corrected map[string]any) string {
	key := strings.TrimSpace(column)
	switch key {
	case "group_name":
		return group.Name
	case "image_id":
		return image.ID.String()
	case "original_filename":
		return image.OriginalFilename
	case "receipt_type":
		if image.ReceiptType != nil {
			return *image.ReceiptType
		}
		if extraction != nil && extraction.ReceiptType != nil {
			return *extraction.ReceiptType
		}
		return ""
	case "ocr_status":
		return image.OCRStatus
	case "overall_confidence":
		if image.OverallConfidence != nil {
			return fmt.Sprintf("%.4f", *image.OverallConfidence)
		}
		return ""
	default:
		if corrected != nil {
			if value, ok := corrected[key]; ok {
				return fmt.Sprint(value)
			}
		}
		if field, ok := fields[key]; ok {
			return fmt.Sprint(field.Value)
		}
		return ""
	}
}
