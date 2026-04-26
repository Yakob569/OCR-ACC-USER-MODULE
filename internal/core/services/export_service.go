package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

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
	if len(selectedColumns) == 0 {
		selectedColumns = []string{"group_name", "original_filename", "receipt_type", "ocr_status", "overall_confidence"}
	}

	images, err := s.imageRepo.ListByGroup(ctx, userID, groupID, 10000, 0)
	if err != nil {
		return nil, err
	}

	exportID := uuid.New()
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)
	if err := writer.Write(selectedColumns); err != nil {
		return nil, err
	}

	for _, image := range images {
		extraction, _ := s.extractRepo.GetByReceiptImageID(ctx, image.ID)
		review, _ := s.reviewRepo.GetByReceiptImageID(ctx, image.ID)

		fields := map[string]fieldValueForExport{}
		if extraction != nil && len(extraction.FieldsJSON) > 0 {
			_ = json.Unmarshal(extraction.FieldsJSON, &fields)
		}

		corrected := map[string]any{}
		if includeCorrectedValues && review != nil && len(review.CorrectedFieldsJSON) > 0 {
			_ = json.Unmarshal(review.CorrectedFieldsJSON, &corrected)
		}

		row := make([]string, 0, len(selectedColumns))
		for _, column := range selectedColumns {
			row = append(row, exportColumnValue(column, group, &image, extraction, fields, corrected))
		}
		if err := writer.Write(row); err != nil {
			return nil, err
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
