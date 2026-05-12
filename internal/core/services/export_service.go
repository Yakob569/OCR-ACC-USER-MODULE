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

			fields := map[string]fieldValueForExport{}
			var totals struct {
				Subtotal  *float64 `json:"subtotal"`
				TaxTotal  *float64 `json:"tax_total"`
			}

			if extraction != nil && len(extraction.FieldsJSON) > 0 {
				_ = json.Unmarshal(extraction.FieldsJSON, &fields)
				// Try to unmarshal totals from the same JSON
				var rawFields map[string]any
				if err := json.Unmarshal(extraction.FieldsJSON, &rawFields); err == nil {
					if t, ok := rawFields["totals"].(map[string]any); ok {
						if sub, ok := t["subtotal"].(map[string]any)["value"].(float64); ok { totals.Subtotal = &sub }
						if tax, ok := t["tax_total"].(map[string]any)["value"].(float64); ok { totals.TaxTotal = &tax }
					}
				}
			}

			var items []map[string]any
			if extraction != nil && len(extraction.ItemsJSON) > 0 {
				_ = json.Unmarshal(extraction.ItemsJSON, &items)
			}

			if len(items) == 0 {
				items = []map[string]any{{}}
			}

			// Calculate sum of line totals for proportional VAT
			var totalLineTotal float64
			for _, item := range items {
				totalLineTotal += s.toFloat(item["line_total"])
			}

			var corrected map[string]any
			if s.reviewRepo != nil {
				review, err := s.reviewRepo.GetByReceiptImageID(ctx, img.ID)
				if err == nil && review != nil && len(review.CorrectedFieldsJSON) > 0 {
					_ = json.Unmarshal(review.CorrectedFieldsJSON, &corrected)
				}
			}

			var imageRows [][]string
			for _, item := range items {
				imageRows = append(imageRows, s.buildVatPurchaseRow(group, &img, extraction, fields, totals.Subtotal, totals.TaxTotal, totalLineTotal, corrected, item))
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
		"VAT \nCATEGORY", "CALENDAR\n TYPE", "Types \nOf purchase", "TIN.", "Seller Name ",
		"Date Of Purchase", "MRC Number", "Vat receipt number", "Description.",
		"Unit of \nMeasure ", "Quantity.", "Unit Price.", "Total value", "vat", "value \nAfter vat",
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
	subtotal *float64,
	taxTotal *float64,
	totalLineTotal float64,
	corrected map[string]any,
	item map[string]any,
) []string {
	row := make([]string, 18)

	// 1. VAT CATEGORY (G=GOODS;S=SERVICES) - Always G
	row[0] = "G"

	// 2. CALENDAR TYPE (E=ETHIOPIAN;G=GREGORIAN) - Forced to G
	row[1] = "G"

	// 3. Types of purchase (1-6) - Dynamic from review stage, default to 5
	if val, ok := corrected["types_of_purchase"]; ok {
		row[2] = fmt.Sprint(val)
	} else {
		row[2] = "5" // Default value
	}

	// 4. TIN
	row[3] = s.getFieldValue(fields, "tin", "merchant")

	// 5. Seller name
	row[4] = s.getFieldValue(fields, "name", "merchant")

	// 6. Date of purchase (dd/mm/yyyy)
	row[5] = s.getFormattedDate(fields)

	// 7. MRC Number (Machine ID)
	row[6] = s.getFieldValue(fields, "machine_id", "transaction")

	// 8. Vat receipt number (FS NO)
	fsNo := s.getFieldValue(fields, "fs_number", "transaction")
	if fsNo == "" {
		fsNo = s.getFieldValue(fields, "invoice_number", "transaction")
	}
	row[7] = fsNo

	// 9. Description
	row[8] = fmt.Sprint(item["description"])
	if row[8] == "<nil>" { row[8] = "" }

	// 10. Unit of Measure - From review stage, fallback to logic
	if val, ok := corrected["unit_of_measurement"]; ok {
		row[9] = fmt.Sprint(val)
	} else {
		row[9] = "7" 
		desc := strings.ToLower(row[8])
		if strings.Contains(desc, "transport") || strings.Contains(desc, "lime") || strings.Contains(desc, "dolomite") || strings.Contains(desc, "utility") || strings.Contains(desc, "marble") {
			row[9] = "9"
		}
	}

	// 11. Quantity
	row[10] = fmt.Sprint(item["quantity"])
	if row[10] == "<nil>" { row[10] = "" }

	// 12. Unit Price
	row[11] = fmt.Sprint(item["unit_price"])
	if row[11] == "<nil>" { row[11] = "" }

	// 13. Total value (Line Total)
	lt := s.toFloat(item["line_total"])
	row[12] = fmt.Sprintf("%.2f", lt)
	if row[12] == "0.00" { row[12] = "" }

	// 14. VAT (Proportional distribution)
	itemVAT := 0.0
	if subtotal != nil && *subtotal > 0 && taxTotal != nil && totalLineTotal > 0 {
		itemVAT = (lt / totalLineTotal) * (*taxTotal)
	} else {
		// Fallback to item-specific tax if total subtotal not available
		itemVAT = s.toFloat(item["tax_amount"])
	}
	row[13] = fmt.Sprintf("%.2f", itemVAT)

	// 15. Value after VAT
	row[14] = fmt.Sprintf("%.2f", lt+itemVAT)

	// 16-18. Empty columns as per template
	row[15] = ""
	row[16] = ""
	row[17] = ""

	return row
}

func (s *groupExportService) getCalendarType(fields map[string]fieldValueForExport) string {
	dateStr := s.getFieldValue(fields, "date", "transaction")
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

func (s *groupExportService) isTaxable(fields map[string]fieldValueForExport, item map[string]any) bool {
	tax := s.toFloat(item["tax_amount"])
	if tax > 0 {
		return true
	}
	// Also check totals if item tax is missing
	totalTax := s.toFloat(s.getFieldValue(fields, "tax_total", "totals"))
	return totalTax > 0
}

func (s *groupExportService) getFormattedDate(fields map[string]fieldValueForExport) string {
	return s.getFieldValue(fields, "date", "transaction")
}

func (s *groupExportService) getFieldValue(fields map[string]fieldValueForExport, key string, section string) string {
	// 1. Check fields map (which maps to extraction.FieldsJSON)
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
