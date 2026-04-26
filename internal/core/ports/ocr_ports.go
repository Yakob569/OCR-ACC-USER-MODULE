package ports

import (
	"context"
	"io"
	"time"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/google/uuid"
)

type ReceiptGroupRepository interface {
	Create(ctx context.Context, input domain.CreateReceiptGroupInput) (*domain.ReceiptGroup, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.ReceiptGroup, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ReceiptGroup, error)
	GetByUserAndID(ctx context.Context, userID, id uuid.UUID) (*domain.ReceiptGroup, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	IncrementImageCounters(ctx context.Context, id uuid.UUID, total, queued, processing, completed, failed, reviewed, exports int) error
}

type ReceiptImageRepository interface {
	CreateMany(ctx context.Context, inputs []domain.ReceiptImageUploadInput) ([]domain.ReceiptImage, error)
	ListByGroup(ctx context.Context, userID, groupID uuid.UUID, limit, offset int) ([]domain.ReceiptImage, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ReceiptImage, error)
	GetByUserAndID(ctx context.Context, userID, id uuid.UUID) (*domain.ReceiptImage, error)
	UpdateStatuses(ctx context.Context, id uuid.UUID, uploadStatus, ocrStatus, reviewStatus string) error
	UpdateProcessingResult(ctx context.Context, id uuid.UUID, status string, receiptType *string, confidence *float64, errorCode, errorMessage *string) error
}

type OCRExtractionRepository interface {
	Upsert(ctx context.Context, extraction *domain.OCRExtraction) error
	GetByReceiptImageID(ctx context.Context, receiptImageID uuid.UUID) (*domain.OCRExtraction, error)
}

type OCRJobRepository interface {
	CreateMany(ctx context.Context, jobs []OCRJobCreateInput) ([]domain.OCRJob, error)
	ListQueued(ctx context.Context, limit int) ([]domain.OCRJob, error)
	ClaimQueued(ctx context.Context, workerID string, limit int) ([]domain.OCRJob, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.OCRJob, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, workerID, errorCode, errorMessage *string) error
	IncrementAttempt(ctx context.Context, id uuid.UUID) error
}

type ObjectStorageService interface {
	UploadReceiptImage(ctx context.Context, userID, groupID, imageID uuid.UUID, filename, contentType string, content io.Reader, contentLength int64) (bucket string, objectKey string, objectURL *string, err error)
	DownloadReceiptImage(ctx context.Context, bucket, objectKey string) ([]byte, error)
}

type OCREngineService interface {
	Extract(ctx context.Context, filename, contentType string, content io.Reader) (*domain.OCRProcessResult, error)
}

type ReceiptGroupService interface {
	CreateGroup(ctx context.Context, input domain.CreateReceiptGroupInput) (*domain.ReceiptGroup, error)
	ListGroups(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.ReceiptGroup, error)
	GetGroup(ctx context.Context, userID, groupID uuid.UUID) (*domain.ReceiptGroup, error)
}

type ReceiptUploadService interface {
	UploadGroupImages(ctx context.Context, groupID, userID uuid.UUID, files []ReceiptFile) (*ReceiptUploadResult, error)
}

type OCRJobService interface {
	ProcessJob(ctx context.Context, jobID uuid.UUID) error
	StartWorkers(ctx context.Context)
}

type DashboardService interface {
	GetSummary(ctx context.Context, userID uuid.UUID) (*domain.DashboardSummary, error)
}

type ReceiptFile struct {
	Filename      string
	ContentType   string
	ContentLength int64
	Bytes         []byte
}

type ReceiptUploadResult struct {
	Images []domain.ReceiptImage
	Jobs   []domain.OCRJob
}

type OCRJobCreateInput struct {
	ReceiptImageID uuid.UUID
	GroupID        uuid.UUID
	UserID         uuid.UUID
	Status         string
	AttemptCount   int
	MaxAttempts    int
	QueuedAt       time.Time
}
