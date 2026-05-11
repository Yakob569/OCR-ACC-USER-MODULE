package services

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
)

// Stubs for testing
type stubGroupRepo struct {
	ports.ReceiptGroupRepository
	getFn func(ctx context.Context, userID, id uuid.UUID) (*domain.ReceiptGroup, error)
	refreshFn func(ctx context.Context, id uuid.UUID) (*domain.ReceiptGroup, error)
}
func (s *stubGroupRepo) GetByUserAndID(ctx context.Context, u, id uuid.UUID) (*domain.ReceiptGroup, error) { return s.getFn(ctx, u, id) }
func (s *stubGroupRepo) RefreshAggregateState(ctx context.Context, id uuid.UUID) (*domain.ReceiptGroup, error) { return s.refreshFn(ctx, id) }

type stubImageRepo struct {
	ports.ReceiptImageRepository
	listFn func(ctx context.Context, u, g uuid.UUID, l, o int) ([]domain.ReceiptImage, error)
}
func (s *stubImageRepo) ListByGroup(ctx context.Context, u, g uuid.UUID, l, o int) ([]domain.ReceiptImage, error) { return s.listFn(ctx, u, g, l, o) }

type stubExtractRepo struct {
	ports.OCRExtractionRepository
	getFn func(ctx context.Context, id uuid.UUID) (*domain.OCRExtraction, error)
}
func (s *stubExtractRepo) GetByReceiptImageID(ctx context.Context, id uuid.UUID) (*domain.OCRExtraction, error) { return s.getFn(ctx, id) }

type stubExportRepo struct {
	ports.GroupExportRepository
	createFn func(ctx context.Context, in domain.CreateGroupExportInput) (*domain.GroupExport, error)
}
func (s *stubExportRepo) Create(ctx context.Context, in domain.CreateGroupExportInput) (*domain.GroupExport, error) { return s.createFn(ctx, in) }

type stubStorage struct {
	ports.ObjectStorageService
	uploadFn func(ctx context.Context, u, g, e uuid.UUID, c []byte) (string, string, *string, error)
}
func (s *stubStorage) UploadGroupExport(ctx context.Context, u, g, e uuid.UUID, c []byte) (string, string, *string, error) { return s.uploadFn(ctx, u, g, e, c) }

func TestCreateCSVExport_VatTemplate(t *testing.T) {
	userID := uuid.New()
	groupID := uuid.New()
	
	groupRepo := &stubGroupRepo{
		getFn: func(ctx context.Context, u, id uuid.UUID) (*domain.ReceiptGroup, error) {
			return &domain.ReceiptGroup{ID: id, Name: "Test Group"}, nil
		},
		refreshFn: func(ctx context.Context, id uuid.UUID) (*domain.ReceiptGroup, error) { return nil, nil },
	}
	
	imageRepo := &stubImageRepo{
		listFn: func(ctx context.Context, u, g uuid.UUID, l, o int) ([]domain.ReceiptImage, error) {
			return []domain.ReceiptImage{
				{ID: uuid.New(), OriginalFilename: "receipt1.jpg", OCRStatus: domain.OCRStatusCompleted},
			}, nil
		},
	}
	
	// Create an extraction with 2 items to test expansion
	items := []map[string]any{
		{"description": "Burger", "quantity": 2, "unit_price": 100, "line_total": 200, "tax_amount": 30},
		{"description": "Soda", "quantity": 1, "unit_price": 50, "line_total": 50, "tax_amount": 0},
	}
	itemsJSON, _ := json.Marshal(items)
	
	fields := map[string]any{
		"tin":        map[string]any{"value": "12345"},
		"name":       map[string]any{"value": "King Burger"},
		"date":       map[string]any{"value": "15/05/2026"}, // 2026 should be Gregorian
		"machine_id": map[string]any{"value": "MAC001"},
		"fs_number":  map[string]any{"value": "FS789"},
	}
	fieldsJSON, _ := json.Marshal(fields)

	extractRepo := &stubExtractRepo{
		getFn: func(ctx context.Context, id uuid.UUID) (*domain.OCRExtraction, error) {
			return &domain.OCRExtraction{
				FieldsJSON: fieldsJSON,
				ItemsJSON:  itemsJSON,
			}, nil
		},
	}
	
	var capturedCSV string
	storage := &stubStorage{
		uploadFn: func(ctx context.Context, u, g, e uuid.UUID, c []byte) (string, string, *string, error) {
			capturedCSV = string(c)
			url := "http://storage/export.csv"
			return "bucket", "key", &url, nil
		},
	}
	
	exportRepo := &stubExportRepo{
		createFn: func(ctx context.Context, in domain.CreateGroupExportInput) (*domain.GroupExport, error) {
			return &domain.GroupExport{ID: uuid.New()}, nil
		},
	}
	
	svc := NewGroupExportService(groupRepo, imageRepo, extractRepo, nil, exportRepo, storage)
	
	_, err := svc.CreateCSVExport(context.Background(), userID, groupID, nil, false)
	if err != nil {
		t.Fatalf("Failed to create export: %v", err)
	}
	
	lines := strings.Split(strings.TrimSpace(capturedCSV), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines (header + 2 items), got %d", len(lines))
	}
	
	// Check first item (Taxable)
	// G, G, 5, 12345, King Burger, 15/05/2026, MAC001, FS789, Burger, 7, 2, 100, 200, 30, 230.00, , ,
	cols1 := strings.Split(lines[1], ",")
	if cols1[0] != "G" { t.Errorf("Expected G, got %s", cols1[0]) }
	if cols1[1] != "G" { t.Errorf("Expected Gregorian (G), got %s", cols1[1]) }
	if cols1[2] != "5" { t.Errorf("Expected Type 5, got %s", cols1[2]) }
	if cols1[7] != "FS789" { t.Errorf("Expected FS NO FS789, got %s", cols1[7]) }
	if cols1[8] != "Burger" { t.Errorf("Expected Description Burger, got %s", cols1[8]) }
	if cols1[9] != "7" { t.Errorf("Expected Unit of Measure 7, got %s", cols1[9]) }
	if cols1[13] != "30" { t.Errorf("Expected Tax 30, got %s", cols1[13]) }
	if cols1[14] != "230.00" { t.Errorf("Expected Value After Vat 230.00, got %s", cols1[14]) }

	// Check second item (Exempt)
	// G, G, 5, 12345, King Burger, 15/05/2026, MAC001, FS789, Soda, 7, 1, 50, 50, 0, 50.00, , ,
	cols2 := strings.Split(lines[2], ",")
	if cols2[2] != "5" { t.Errorf("Expected Type 5, got %s", cols2[2]) }
	if cols2[8] != "Soda" { t.Errorf("Expected Description Soda, got %s", cols2[8]) }
	if cols2[13] != "0" { t.Errorf("Expected Tax 0, got %s", cols2[13]) }
}

func TestCalendarDetection(t *testing.T) {
	svc := &groupExportService{}
	
	// Gregorian test
	fieldsG := map[string]fieldValueForExport{
		"date": {Value: "15/05/2026"},
	}
	if cal := svc.getCalendarType(fieldsG, nil); cal != "G" {
		t.Errorf("Expected G for 2026, got %s", cal)
	}
	
	// Ethiopian test (assuming current year is 2026+)
	fieldsE := map[string]fieldValueForExport{
		"date": {Value: "15/05/2018"},
	}
	if cal := svc.getCalendarType(fieldsE, nil); cal != "E" {
		t.Errorf("Expected E for 2018, got %s", cal)
	}
}
