package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	GroupStatusDraft                 = "draft"
	GroupStatusUploading             = "uploading"
	GroupStatusQueued                = "queued"
	GroupStatusProcessing            = "processing"
	GroupStatusCompleted             = "completed"
	GroupStatusCompletedWithFailures = "completed_with_failures"
	GroupStatusFailed                = "failed"
	GroupStatusArchived              = "archived"

	UploadStatusPending    = "pending"
	UploadStatusUploaded   = "uploaded"
	UploadStatusUploadFail = "upload_failed"

	OCRStatusQueued      = "queued"
	OCRStatusProcessing  = "processing"
	OCRStatusCompleted   = "completed"
	OCRStatusFailed      = "failed"
	OCRStatusNeedsReview = "needs_review"

	ReviewStatusPending  = "pending"
	ReviewStatusReviewed = "reviewed"
	ReviewStatusAccepted = "accepted"
	ReviewStatusRejected = "rejected"

	JobStatusQueued     = "queued"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
	JobStatusRetrying   = "retrying"
	JobStatusCancelled  = "cancelled"
)

type ReceiptGroup struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	Name             string    `json:"name"`
	Description      *string   `json:"description"`
	Status           string    `json:"status"`
	TotalImages      int       `json:"total_images"`
	QueuedImages     int       `json:"queued_images"`
	ProcessingImages int       `json:"processing_images"`
	CompletedImages  int       `json:"completed_images"`
	FailedImages     int       `json:"failed_images"`
	ReviewedImages   int       `json:"reviewed_images"`
	ExportCount      int       `json:"export_count"`
	DeletedAt        *time.Time `json:"deleted_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ReceiptImage struct {
	ID                uuid.UUID  `json:"id"`
	GroupID           uuid.UUID  `json:"group_id"`
	UserID            uuid.UUID  `json:"user_id"`
	OriginalFilename  string     `json:"original_filename"`
	MIMEType          string     `json:"mime_type"`
	FileSizeBytes     int64      `json:"file_size_bytes"`
	ChecksumSHA256    string     `json:"checksum_sha256"`
	StorageBucket     string     `json:"storage_bucket"`
	StorageObjectKey  string     `json:"storage_object_key"`
	StorageURL        *string    `json:"storage_url"`
	UploadStatus      string     `json:"upload_status"`
	OCRStatus         string     `json:"ocr_status"`
	ReviewStatus      string     `json:"review_status"`
	OCRAttemptCount   int        `json:"ocr_attempt_count"`
	LastErrorCode     *string    `json:"last_error_code"`
	LastErrorMessage  *string    `json:"last_error_message"`
	ReceiptType       *string    `json:"receipt_type"`
	OverallConfidence *float64   `json:"overall_confidence"`
	ProcessedAt       *time.Time `json:"processed_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type OCRExtraction struct {
	ID               uuid.UUID `json:"id"`
	ReceiptImageID   uuid.UUID `json:"receipt_image_id"`
	Success          bool      `json:"success"`
	ReceiptType      *string   `json:"receipt_type"`
	FieldsJSON       []byte    `json:"fields_json"`
	ItemsJSON        []byte    `json:"items_json"`
	WarningsJSON     []byte    `json:"warnings_json"`
	RawText          *string   `json:"raw_text"`
	DebugJSON        []byte    `json:"debug_json"`
	OCREngineURL     *string   `json:"ocr_engine_url"`
	OCREngineVersion *string   `json:"ocr_engine_version"`
	PipelineVersion  *string   `json:"pipeline_version"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type OCRJob struct {
	ID             uuid.UUID  `json:"id"`
	ReceiptImageID uuid.UUID  `json:"receipt_image_id"`
	GroupID        uuid.UUID  `json:"group_id"`
	UserID         uuid.UUID  `json:"user_id"`
	Status         string     `json:"status"`
	AttemptCount   int        `json:"attempt_count"`
	MaxAttempts    int        `json:"max_attempts"`
	QueuedAt       time.Time  `json:"queued_at"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	WorkerID       *string    `json:"worker_id"`
	ErrorCode      *string    `json:"error_code"`
	ErrorMessage   *string    `json:"error_message"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ReceiptReview struct {
	ID                  uuid.UUID `json:"id"`
	ReceiptImageID      uuid.UUID `json:"receipt_image_id"`
	ReviewedByUserID    uuid.UUID `json:"reviewed_by_user_id"`
	QualityLabel        string    `json:"quality_label"`
	IsAccepted          bool      `json:"is_accepted"`
	CorrectedFieldsJSON []byte    `json:"corrected_fields_json"`
	ReviewNotes         *string   `json:"review_notes"`
	ReviewedAt          time.Time `json:"reviewed_at"`
	CreatedAt           time.Time `json:"created_at"`
}

type GroupExport struct {
	ID                  uuid.UUID `json:"id"`
	GroupID             uuid.UUID `json:"group_id"`
	ExportedByUserID    uuid.UUID `json:"exported_by_user_id"`
	Format              string    `json:"format"`
	SelectedColumnsJSON []byte    `json:"selected_columns_json"`
	RowCount            int       `json:"row_count"`
	StorageBucket       *string   `json:"storage_bucket"`
	StorageObjectKey    *string   `json:"storage_object_key"`
	StorageURL          *string   `json:"storage_url"`
	CreatedAt           time.Time `json:"created_at"`
}

type DashboardSummary struct {
	TotalGroups          int            `json:"total_groups"`
	TotalScans           int            `json:"total_scans"`
	SuccessfulScans      int            `json:"successful_scans"`
	FailedScans          int            `json:"failed_scans"`
	NeedsReviewScans     int            `json:"needs_review_scans"`
	AverageConfidence    *float64       `json:"average_confidence"`
	AcceptedAccuracyRate *float64       `json:"accepted_accuracy_rate"`
	RecentGroups         []ReceiptGroup `json:"recent_groups"`
	RecentImages         []ReceiptImage `json:"recent_images"`
}

type CreateReceiptGroupInput struct {
	UserID      uuid.UUID
	Name        string
	Description string
}

type ReceiptImageUploadInput struct {
	ID               uuid.UUID
	GroupID          uuid.UUID
	UserID           uuid.UUID
	OriginalFilename string
	MIMEType         string
	FileSizeBytes    int64
	ChecksumSHA256   string
	StorageBucket    string
	StorageObjectKey string
	StorageURL       string
}

type OCRProcessResult struct {
	Extraction  *OCRExtraction
	OCRStatus   string
	ReceiptType *string
	Confidence  *float64
}

type SubmitReceiptReviewInput struct {
	ReceiptImageID      uuid.UUID
	ReviewedByUserID    uuid.UUID
	QualityLabel        string
	IsAccepted          bool
	CorrectedFieldsJSON []byte
	ReviewNotes         string
}

type CreateGroupExportInput struct {
	GroupID             uuid.UUID
	ExportedByUserID    uuid.UUID
	SelectedColumnsJSON []byte
	RowCount            int
	StorageBucket       string
	StorageObjectKey    string
	StorageURL          string
}
