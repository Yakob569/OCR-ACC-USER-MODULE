package main
import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/google/uuid"
	"os"
)
func main() {
	conn, _ := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	// Image ID that we know HAS an extraction
	imageID, _ := uuid.Parse("20383200-6dde-4c96-aa3a-40adce7a67a1")
	
	// Query as implemented in OCRExtractionRepository
	query := "SELECT id, receipt_image_id, success, receipt_type, fields_json, items_json, warnings_json, raw_text, debug_json, ocr_engine_url, ocr_engine_version, pipeline_version, created_at, updated_at FROM receipt_extractions WHERE receipt_image_id = $1"
	
	var id, receiptImageID uuid.UUID
    // ... just scan the first few fields
	var success bool
	err := conn.QueryRow(context.Background(), query, imageID).Scan(&id, &receiptImageID, &success, new(interface{}), new(interface{}), new(interface{}), new(interface{}), new(interface{}), new(interface{}), new(interface{}), new(interface{}), new(interface{}), new(interface{}), new(interface{}))
	fmt.Printf("GetByReceiptImageID check result: %v\n", err)
}
