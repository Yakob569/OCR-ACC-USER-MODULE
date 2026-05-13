package main
import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
	"github.com/google/uuid"
    "os"
)
type OCRExtraction struct {
    ID               uuid.UUID 
    ReceiptImageID   uuid.UUID 
    Success          bool      
    ReceiptType      *string   
    FieldsJSON       []byte    
    ItemsJSON        []byte    
    WarningsJSON     []byte    
    RawText          *string   
    DebugJSON        []byte    
    OCREngineURL     *string   
    OCREngineVersion *string   
    PipelineVersion  *string   
    CreatedAt        interface{}
    UpdatedAt        interface{}
}
func main() {
	conn, _ := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	imageID, _ := uuid.Parse("20383200-6dde-4c96-aa3a-40adce7a67a1")
	query := "SELECT id, receipt_image_id, success, receipt_type, fields_json, items_json, warnings_json, raw_text, debug_json, ocr_engine_url, ocr_engine_version, pipeline_version, created_at, updated_at FROM receipt_extractions WHERE receipt_image_id = $1"
	
	var ext OCRExtraction
	var receiptType, rawText, ocrEngineURL, ocrEngineVersion, pipelineVersion pgtype.Text
	var debugJSON []byte

	err := conn.QueryRow(context.Background(), query, imageID).Scan(
		&ext.ID, &ext.ReceiptImageID, &ext.Success, &receiptType, &ext.FieldsJSON, &ext.ItemsJSON, &ext.WarningsJSON,
		&rawText, &debugJSON, &ocrEngineURL, &ocrEngineVersion, &pipelineVersion,
		&ext.CreatedAt, &ext.UpdatedAt,
	)
    ext.ReceiptType = nil
    if receiptType.Valid {
        v := receiptType.String
        ext.ReceiptType = &v
    }
    fmt.Printf("Scan err: %v, ReceiptType: %v\n", err, ext.ReceiptType)
}
