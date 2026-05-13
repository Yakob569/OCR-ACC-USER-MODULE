package main
import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
	"github.com/google/uuid"
    "os"
)
func main() {
	conn, _ := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	imageID, _ := uuid.Parse("20383200-6dde-4c96-aa3a-40adce7a67a1")
	query := "SELECT id, receipt_image_id, success, receipt_type, fields_json, items_json, warnings_json, raw_text, debug_json, ocr_engine_url, ocr_engine_version, pipeline_version, created_at, updated_at FROM receipt_extractions WHERE receipt_image_id = $1"
	
	var id, receiptImageID uuid.UUID
	var success bool
	var receiptType, rawText, ocrEngineURL, ocrEngineVersion, pipelineVersion pgtype.Text
	var fieldsJSON, itemsJSON, warningsJSON []byte
	var debugJSON []byte
	var createdAt, updatedAt interface{}

	err := conn.QueryRow(context.Background(), query, imageID).Scan(
		&id, &receiptImageID, &success, &receiptType, &fieldsJSON, &itemsJSON, &warningsJSON,
		&rawText, &debugJSON, &ocrEngineURL, &ocrEngineVersion, &pipelineVersion,
		&createdAt, &updatedAt,
	)
	fmt.Printf("Scan err: %v\n", err)
}
