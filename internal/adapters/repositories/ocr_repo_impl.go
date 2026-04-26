package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type receiptGroupRepo struct {
	db *pgxpool.Pool
}

type receiptImageRepo struct {
	db *pgxpool.Pool
}

type ocrExtractionRepo struct {
	db *pgxpool.Pool
}

type ocrJobRepo struct {
	db *pgxpool.Pool
}

type dashboardRepo struct {
	db *pgxpool.Pool
}

func NewReceiptGroupRepository(db *pgxpool.Pool) ports.ReceiptGroupRepository {
	return &receiptGroupRepo{db: db}
}

func NewReceiptImageRepository(db *pgxpool.Pool) ports.ReceiptImageRepository {
	return &receiptImageRepo{db: db}
}

func NewOCRExtractionRepository(db *pgxpool.Pool) ports.OCRExtractionRepository {
	return &ocrExtractionRepo{db: db}
}

func NewOCRJobRepository(db *pgxpool.Pool) ports.OCRJobRepository {
	return &ocrJobRepo{db: db}
}

func NewDashboardRepository(db *pgxpool.Pool) ports.DashboardRepository {
	return &dashboardRepo{db: db}
}

func (r *receiptGroupRepo) Create(ctx context.Context, input domain.CreateReceiptGroupInput) (*domain.ReceiptGroup, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		INSERT INTO receipt_groups (user_id, name, description)
		VALUES ($1, $2, NULLIF($3, ''))
		RETURNING id, user_id, name, description, status, total_images, queued_images, processing_images,
		          completed_images, failed_images, reviewed_images, export_count, created_at, updated_at
	`

	var group domain.ReceiptGroup
	var description pgtype.Text
	err := r.db.QueryRow(ctx, query, input.UserID, input.Name, input.Description).Scan(
		&group.ID,
		&group.UserID,
		&group.Name,
		&description,
		&group.Status,
		&group.TotalImages,
		&group.QueuedImages,
		&group.ProcessingImages,
		&group.CompletedImages,
		&group.FailedImages,
		&group.ReviewedImages,
		&group.ExportCount,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	group.Description = nullableText(description)
	return &group, nil
}

func (r *receiptGroupRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.ReceiptGroup, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		SELECT id, user_id, name, description, status, total_images, queued_images, processing_images,
		       completed_images, failed_images, reviewed_images, export_count, created_at, updated_at
		FROM receipt_groups
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []domain.ReceiptGroup
	for rows.Next() {
		var group domain.ReceiptGroup
		var description pgtype.Text
		if err := rows.Scan(
			&group.ID,
			&group.UserID,
			&group.Name,
			&description,
			&group.Status,
			&group.TotalImages,
			&group.QueuedImages,
			&group.ProcessingImages,
			&group.CompletedImages,
			&group.FailedImages,
			&group.ReviewedImages,
			&group.ExportCount,
			&group.CreatedAt,
			&group.UpdatedAt,
		); err != nil {
			return nil, err
		}
		group.Description = nullableText(description)
		groups = append(groups, group)
	}

	return groups, rows.Err()
}

func (r *receiptGroupRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ReceiptGroup, error) {
	return r.getOne(ctx, `SELECT id, user_id, name, description, status, total_images, queued_images, processing_images,
		completed_images, failed_images, reviewed_images, export_count, created_at, updated_at
		FROM receipt_groups WHERE id = $1`, id)
}

func (r *receiptGroupRepo) GetByUserAndID(ctx context.Context, userID, id uuid.UUID) (*domain.ReceiptGroup, error) {
	return r.getOne(ctx, `SELECT id, user_id, name, description, status, total_images, queued_images, processing_images,
		completed_images, failed_images, reviewed_images, export_count, created_at, updated_at
		FROM receipt_groups WHERE user_id = $1 AND id = $2`, userID, id)
}

func (r *receiptGroupRepo) getOne(ctx context.Context, query string, args ...any) (*domain.ReceiptGroup, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	var group domain.ReceiptGroup
	var description pgtype.Text
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&group.ID,
		&group.UserID,
		&group.Name,
		&description,
		&group.Status,
		&group.TotalImages,
		&group.QueuedImages,
		&group.ProcessingImages,
		&group.CompletedImages,
		&group.FailedImages,
		&group.ReviewedImages,
		&group.ExportCount,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	group.Description = nullableText(description)
	return &group, nil
}

func (r *receiptGroupRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if r.db == nil {
		return errors.New("database connection is not available")
	}

	_, err := r.db.Exec(ctx, `UPDATE receipt_groups SET status = $2 WHERE id = $1`, id, status)
	return err
}

func (r *receiptGroupRepo) IncrementImageCounters(ctx context.Context, id uuid.UUID, total, queued, processing, completed, failed, reviewed, exports int) error {
	if r.db == nil {
		return errors.New("database connection is not available")
	}

	query := `
		UPDATE receipt_groups
		SET total_images = total_images + $2,
		    queued_images = queued_images + $3,
		    processing_images = processing_images + $4,
		    completed_images = completed_images + $5,
		    failed_images = failed_images + $6,
		    reviewed_images = reviewed_images + $7,
		    export_count = export_count + $8
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, total, queued, processing, completed, failed, reviewed, exports)
	return err
}

func (r *receiptImageRepo) CreateMany(ctx context.Context, inputs []domain.ReceiptImageUploadInput) ([]domain.ReceiptImage, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}
	if len(inputs) == 0 {
		return []domain.ReceiptImage{}, nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO receipt_images (
			id, group_id, user_id, original_filename, mime_type, file_size_bytes, checksum_sha256,
			storage_bucket, storage_object_key, storage_url, upload_status, ocr_status, review_status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), $11, $12, $13)
		RETURNING id, group_id, user_id, original_filename, mime_type, file_size_bytes, checksum_sha256,
		          storage_bucket, storage_object_key, storage_url, upload_status, ocr_status, review_status,
		          ocr_attempt_count, last_error_code, last_error_message, receipt_type, overall_confidence,
		          processed_at, created_at, updated_at
	`

	images := make([]domain.ReceiptImage, 0, len(inputs))
	for _, input := range inputs {
		var image domain.ReceiptImage
		var storageURL, lastErrorCode, lastErrorMessage, receiptType pgtype.Text
		var overallConfidence pgtype.Float8
		var processedAt pgtype.Timestamptz

		err := tx.QueryRow(ctx, query,
			input.ID,
			input.GroupID,
			input.UserID,
			input.OriginalFilename,
			input.MIMEType,
			input.FileSizeBytes,
			input.ChecksumSHA256,
			input.StorageBucket,
			input.StorageObjectKey,
			input.StorageURL,
			domain.UploadStatusUploaded,
			domain.OCRStatusQueued,
			domain.ReviewStatusPending,
		).Scan(
			&image.ID,
			&image.GroupID,
			&image.UserID,
			&image.OriginalFilename,
			&image.MIMEType,
			&image.FileSizeBytes,
			&image.ChecksumSHA256,
			&image.StorageBucket,
			&image.StorageObjectKey,
			&storageURL,
			&image.UploadStatus,
			&image.OCRStatus,
			&image.ReviewStatus,
			&image.OCRAttemptCount,
			&lastErrorCode,
			&lastErrorMessage,
			&receiptType,
			&overallConfidence,
			&processedAt,
			&image.CreatedAt,
			&image.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		image.StorageURL = nullableText(storageURL)
		image.LastErrorCode = nullableText(lastErrorCode)
		image.LastErrorMessage = nullableText(lastErrorMessage)
		image.ReceiptType = nullableText(receiptType)
		image.OverallConfidence = nullableFloat(overallConfidence)
		image.ProcessedAt = nullableTime(processedAt)
		images = append(images, image)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return images, nil
}

func (r *receiptImageRepo) ListByGroup(ctx context.Context, userID, groupID uuid.UUID, limit, offset int) ([]domain.ReceiptImage, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		SELECT id, group_id, user_id, original_filename, mime_type, file_size_bytes, checksum_sha256,
		       storage_bucket, storage_object_key, storage_url, upload_status, ocr_status, review_status,
		       ocr_attempt_count, last_error_code, last_error_message, receipt_type, overall_confidence,
		       processed_at, created_at, updated_at
		FROM receipt_images
		WHERE user_id = $1 AND group_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, userID, groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []domain.ReceiptImage
	for rows.Next() {
		image, err := scanReceiptImage(rows)
		if err != nil {
			return nil, err
		}
		images = append(images, *image)
	}

	return images, rows.Err()
}

func (r *receiptImageRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ReceiptImage, error) {
	return r.getOne(ctx, `SELECT id, group_id, user_id, original_filename, mime_type, file_size_bytes, checksum_sha256,
		storage_bucket, storage_object_key, storage_url, upload_status, ocr_status, review_status,
		ocr_attempt_count, last_error_code, last_error_message, receipt_type, overall_confidence,
		processed_at, created_at, updated_at
		FROM receipt_images WHERE id = $1`, id)
}

func (r *receiptImageRepo) GetByUserAndID(ctx context.Context, userID, id uuid.UUID) (*domain.ReceiptImage, error) {
	return r.getOne(ctx, `SELECT id, group_id, user_id, original_filename, mime_type, file_size_bytes, checksum_sha256,
		storage_bucket, storage_object_key, storage_url, upload_status, ocr_status, review_status,
		ocr_attempt_count, last_error_code, last_error_message, receipt_type, overall_confidence,
		processed_at, created_at, updated_at
		FROM receipt_images WHERE user_id = $1 AND id = $2`, userID, id)
}

func (r *receiptImageRepo) getOne(ctx context.Context, query string, args ...any) (*domain.ReceiptImage, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	row := r.db.QueryRow(ctx, query, args...)
	image, err := scanReceiptImage(row)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (r *receiptImageRepo) UpdateStatuses(ctx context.Context, id uuid.UUID, uploadStatus, ocrStatus, reviewStatus string) error {
	if r.db == nil {
		return errors.New("database connection is not available")
	}

	query := `
		UPDATE receipt_images
		SET upload_status = $2,
		    ocr_status = $3,
		    review_status = $4
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, uploadStatus, ocrStatus, reviewStatus)
	return err
}

func (r *receiptImageRepo) UpdateProcessingResult(ctx context.Context, id uuid.UUID, status string, receiptType *string, confidence *float64, errorCode, errorMessage *string) error {
	if r.db == nil {
		return errors.New("database connection is not available")
	}

	query := `
		UPDATE receipt_images
		SET ocr_status = $2,
		    receipt_type = $3,
		    overall_confidence = $4,
		    last_error_code = $5,
		    last_error_message = $6,
		    processed_at = $7
		WHERE id = $1
	`

	var processedAt *time.Time
	if status == domain.OCRStatusCompleted || status == domain.OCRStatusFailed || status == domain.OCRStatusNeedsReview {
		now := time.Now().UTC()
		processedAt = &now
	}

	_, err := r.db.Exec(ctx, query, id, status, receiptType, confidence, errorCode, errorMessage, processedAt)
	return err
}

func (r *ocrExtractionRepo) Upsert(ctx context.Context, extraction *domain.OCRExtraction) error {
	if r.db == nil {
		return errors.New("database connection is not available")
	}

	query := `
		INSERT INTO receipt_extractions (
			receipt_image_id, success, receipt_type, fields_json, items_json, warnings_json,
			raw_text, debug_json, ocr_engine_url, ocr_engine_version, pipeline_version
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (receipt_image_id) DO UPDATE
		SET success = EXCLUDED.success,
		    receipt_type = EXCLUDED.receipt_type,
		    fields_json = EXCLUDED.fields_json,
		    items_json = EXCLUDED.items_json,
		    warnings_json = EXCLUDED.warnings_json,
		    raw_text = EXCLUDED.raw_text,
		    debug_json = EXCLUDED.debug_json,
		    ocr_engine_url = EXCLUDED.ocr_engine_url,
		    ocr_engine_version = EXCLUDED.ocr_engine_version,
		    pipeline_version = EXCLUDED.pipeline_version
	`

	_, err := r.db.Exec(ctx, query,
		extraction.ReceiptImageID,
		extraction.Success,
		extraction.ReceiptType,
		extraction.FieldsJSON,
		extraction.ItemsJSON,
		extraction.WarningsJSON,
		extraction.RawText,
		nullableBytes(extraction.DebugJSON),
		extraction.OCREngineURL,
		extraction.OCREngineVersion,
		extraction.PipelineVersion,
	)
	return err
}

func (r *ocrExtractionRepo) GetByReceiptImageID(ctx context.Context, receiptImageID uuid.UUID) (*domain.OCRExtraction, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		SELECT id, receipt_image_id, success, receipt_type, fields_json, items_json, warnings_json,
		       raw_text, debug_json, ocr_engine_url, ocr_engine_version, pipeline_version,
		       created_at, updated_at
		FROM receipt_extractions
		WHERE receipt_image_id = $1
	`

	var extraction domain.OCRExtraction
	var receiptType, rawText, ocrEngineURL, ocrEngineVersion, pipelineVersion pgtype.Text
	var debugJSON []byte
	err := r.db.QueryRow(ctx, query, receiptImageID).Scan(
		&extraction.ID,
		&extraction.ReceiptImageID,
		&extraction.Success,
		&receiptType,
		&extraction.FieldsJSON,
		&extraction.ItemsJSON,
		&extraction.WarningsJSON,
		&rawText,
		&debugJSON,
		&ocrEngineURL,
		&ocrEngineVersion,
		&pipelineVersion,
		&extraction.CreatedAt,
		&extraction.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	extraction.ReceiptType = nullableText(receiptType)
	extraction.RawText = nullableText(rawText)
	extraction.DebugJSON = debugJSON
	extraction.OCREngineURL = nullableText(ocrEngineURL)
	extraction.OCREngineVersion = nullableText(ocrEngineVersion)
	extraction.PipelineVersion = nullableText(pipelineVersion)
	return &extraction, nil
}

func (r *ocrExtractionRepo) ListByGroup(ctx context.Context, userID, groupID uuid.UUID, limit, offset int) ([]domain.OCRExtraction, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		SELECT re.id, re.receipt_image_id, re.success, re.receipt_type, re.fields_json, re.items_json, re.warnings_json,
		       re.raw_text, re.debug_json, re.ocr_engine_url, re.ocr_engine_version, re.pipeline_version,
		       re.created_at, re.updated_at
		FROM receipt_extractions re
		INNER JOIN receipt_images ri ON ri.id = re.receipt_image_id
		WHERE ri.user_id = $1 AND ri.group_id = $2
		ORDER BY re.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, userID, groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var extractions []domain.OCRExtraction
	for rows.Next() {
		extraction, err := scanOCRExtraction(rows)
		if err != nil {
			return nil, err
		}
		extractions = append(extractions, *extraction)
	}
	return extractions, rows.Err()
}

func (r *ocrJobRepo) CreateMany(ctx context.Context, jobs []ports.OCRJobCreateInput) ([]domain.OCRJob, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}
	if len(jobs) == 0 {
		return []domain.OCRJob{}, nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO ocr_jobs (
			receipt_image_id, group_id, user_id, status, attempt_count, max_attempts,
			queued_at, started_at, finished_at, worker_id, error_code, error_message
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, receipt_image_id, group_id, user_id, status, attempt_count, max_attempts,
		          queued_at, started_at, finished_at, worker_id, error_code, error_message,
		          created_at, updated_at
	`

	createdJobs := make([]domain.OCRJob, 0, len(jobs))
	for _, job := range jobs {
		var created domain.OCRJob
		var startedAt, finishedAt pgtype.Timestamptz
		var workerID, errorCode, errorMessage pgtype.Text

		err := tx.QueryRow(ctx, query,
			job.ReceiptImageID,
			job.GroupID,
			job.UserID,
			job.Status,
			job.AttemptCount,
			job.MaxAttempts,
			job.QueuedAt,
			nil,
			nil,
			nil,
			nil,
			nil,
		).Scan(
			&created.ID,
			&created.ReceiptImageID,
			&created.GroupID,
			&created.UserID,
			&created.Status,
			&created.AttemptCount,
			&created.MaxAttempts,
			&created.QueuedAt,
			&startedAt,
			&finishedAt,
			&workerID,
			&errorCode,
			&errorMessage,
			&created.CreatedAt,
			&created.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		created.StartedAt = nullableTime(startedAt)
		created.FinishedAt = nullableTime(finishedAt)
		created.WorkerID = nullableText(workerID)
		created.ErrorCode = nullableText(errorCode)
		created.ErrorMessage = nullableText(errorMessage)
		createdJobs = append(createdJobs, created)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return createdJobs, nil
}

func (r *ocrJobRepo) ListQueued(ctx context.Context, limit int) ([]domain.OCRJob, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		SELECT id, receipt_image_id, group_id, user_id, status, attempt_count, max_attempts,
		       queued_at, started_at, finished_at, worker_id, error_code, error_message,
		       created_at, updated_at
		FROM ocr_jobs
		WHERE status IN ($1, $2)
		ORDER BY queued_at ASC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, domain.JobStatusQueued, domain.JobStatusRetrying, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.OCRJob
	for rows.Next() {
		job, err := scanOCRJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, *job)
	}

	return jobs, rows.Err()
}

func (r *ocrJobRepo) ClaimQueued(ctx context.Context, workerID string, limit int) ([]domain.OCRJob, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}
	if limit <= 0 {
		return []domain.OCRJob{}, nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		WITH claimed AS (
			SELECT id
			FROM ocr_jobs
			WHERE status IN ($1, $2)
			ORDER BY queued_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT $3
		)
		UPDATE ocr_jobs
		SET status = $4,
		    worker_id = $5,
		    started_at = NOW()
		WHERE id IN (SELECT id FROM claimed)
		RETURNING id, receipt_image_id, group_id, user_id, status, attempt_count, max_attempts,
		          queued_at, started_at, finished_at, worker_id, error_code, error_message,
		          created_at, updated_at
	`

	rows, err := tx.Query(ctx, query,
		domain.JobStatusQueued,
		domain.JobStatusRetrying,
		limit,
		domain.JobStatusProcessing,
		workerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.OCRJob
	for rows.Next() {
		job, err := scanOCRJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, *job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *ocrJobRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.OCRJob, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	query := `
		SELECT id, receipt_image_id, group_id, user_id, status, attempt_count, max_attempts,
		       queued_at, started_at, finished_at, worker_id, error_code, error_message,
		       created_at, updated_at
		FROM ocr_jobs
		WHERE id = $1
	`

	return scanOCRJob(r.db.QueryRow(ctx, query, id))
}

func (r *ocrJobRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, workerID, errorCode, errorMessage *string) error {
	if r.db == nil {
		return errors.New("database connection is not available")
	}

	var startedAt *time.Time
	var finishedAt *time.Time
	now := time.Now().UTC()

	if status == domain.JobStatusProcessing {
		startedAt = &now
	}
	if status == domain.JobStatusCompleted || status == domain.JobStatusFailed || status == domain.JobStatusCancelled {
		finishedAt = &now
	}

	query := `
		UPDATE ocr_jobs
		SET status = $2,
		    worker_id = $3,
		    error_code = $4,
		    error_message = $5,
		    started_at = COALESCE($6, started_at),
		    finished_at = COALESCE($7, finished_at)
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, status, workerID, errorCode, errorMessage, startedAt, finishedAt)
	return err
}

func (r *ocrJobRepo) IncrementAttempt(ctx context.Context, id uuid.UUID) error {
	if r.db == nil {
		return errors.New("database connection is not available")
	}

	_, err := r.db.Exec(ctx, `UPDATE ocr_jobs SET attempt_count = attempt_count + 1 WHERE id = $1`, id)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanReceiptImage(row scanner) (*domain.ReceiptImage, error) {
	var image domain.ReceiptImage
	var storageURL, lastErrorCode, lastErrorMessage, receiptType pgtype.Text
	var overallConfidence pgtype.Float8
	var processedAt pgtype.Timestamptz

	err := row.Scan(
		&image.ID,
		&image.GroupID,
		&image.UserID,
		&image.OriginalFilename,
		&image.MIMEType,
		&image.FileSizeBytes,
		&image.ChecksumSHA256,
		&image.StorageBucket,
		&image.StorageObjectKey,
		&storageURL,
		&image.UploadStatus,
		&image.OCRStatus,
		&image.ReviewStatus,
		&image.OCRAttemptCount,
		&lastErrorCode,
		&lastErrorMessage,
		&receiptType,
		&overallConfidence,
		&processedAt,
		&image.CreatedAt,
		&image.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	image.StorageURL = nullableText(storageURL)
	image.LastErrorCode = nullableText(lastErrorCode)
	image.LastErrorMessage = nullableText(lastErrorMessage)
	image.ReceiptType = nullableText(receiptType)
	image.OverallConfidence = nullableFloat(overallConfidence)
	image.ProcessedAt = nullableTime(processedAt)
	return &image, nil
}

func scanOCRJob(row scanner) (*domain.OCRJob, error) {
	var job domain.OCRJob
	var startedAt, finishedAt pgtype.Timestamptz
	var workerID, errorCode, errorMessage pgtype.Text

	err := row.Scan(
		&job.ID,
		&job.ReceiptImageID,
		&job.GroupID,
		&job.UserID,
		&job.Status,
		&job.AttemptCount,
		&job.MaxAttempts,
		&job.QueuedAt,
		&startedAt,
		&finishedAt,
		&workerID,
		&errorCode,
		&errorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	job.StartedAt = nullableTime(startedAt)
	job.FinishedAt = nullableTime(finishedAt)
	job.WorkerID = nullableText(workerID)
	job.ErrorCode = nullableText(errorCode)
	job.ErrorMessage = nullableText(errorMessage)
	return &job, nil
}

func scanOCRExtraction(row scanner) (*domain.OCRExtraction, error) {
	var extraction domain.OCRExtraction
	var receiptType, rawText, ocrEngineURL, ocrEngineVersion, pipelineVersion pgtype.Text
	var debugJSON []byte

	err := row.Scan(
		&extraction.ID,
		&extraction.ReceiptImageID,
		&extraction.Success,
		&receiptType,
		&extraction.FieldsJSON,
		&extraction.ItemsJSON,
		&extraction.WarningsJSON,
		&rawText,
		&debugJSON,
		&ocrEngineURL,
		&ocrEngineVersion,
		&pipelineVersion,
		&extraction.CreatedAt,
		&extraction.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	extraction.ReceiptType = nullableText(receiptType)
	extraction.RawText = nullableText(rawText)
	extraction.DebugJSON = debugJSON
	extraction.OCREngineURL = nullableText(ocrEngineURL)
	extraction.OCREngineVersion = nullableText(ocrEngineVersion)
	extraction.PipelineVersion = nullableText(pipelineVersion)
	return &extraction, nil
}

func nullableText(value pgtype.Text) *string {
	if value.Valid {
		return &value.String
	}
	return nil
}

func nullableFloat(value pgtype.Float8) *float64 {
	if value.Valid {
		return &value.Float64
	}
	return nil
}

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if value.Valid {
		return &value.Time
	}
	return nil
}

func nullableBytes(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	return value
}

var _ ports.ReceiptGroupRepository = (*receiptGroupRepo)(nil)
var _ ports.ReceiptImageRepository = (*receiptImageRepo)(nil)
var _ ports.OCRExtractionRepository = (*ocrExtractionRepo)(nil)
var _ ports.OCRJobRepository = (*ocrJobRepo)(nil)
var _ ports.DashboardRepository = (*dashboardRepo)(nil)

func (r *dashboardRepo) GetSummary(ctx context.Context, userID uuid.UUID) (*domain.DashboardSummary, error) {
	if r.db == nil {
		return nil, errors.New("database connection is not available")
	}

	summary := &domain.DashboardSummary{}

	if err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM receipt_groups WHERE user_id = $1),
			(SELECT COUNT(*) FROM receipt_images WHERE user_id = $1),
			(SELECT COUNT(*) FROM receipt_images WHERE user_id = $1 AND ocr_status = $2),
			(SELECT COUNT(*) FROM receipt_images WHERE user_id = $1 AND ocr_status = $3),
			(SELECT COUNT(*) FROM receipt_images WHERE user_id = $1 AND ocr_status = $4)
	`, userID, domain.OCRStatusCompleted, domain.OCRStatusFailed, domain.OCRStatusNeedsReview).Scan(
		&summary.TotalGroups,
		&summary.TotalScans,
		&summary.SuccessfulScans,
		&summary.FailedScans,
		&summary.NeedsReviewScans,
	); err != nil {
		return nil, err
	}

	var averageConfidence pgtype.Float8
	var acceptedAccuracy pgtype.Float8
	if err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT AVG(overall_confidence) FROM receipt_images WHERE user_id = $1 AND overall_confidence IS NOT NULL),
			(
				SELECT AVG(CASE WHEN quality_label = 'accurate' THEN 1.0 ELSE 0.0 END)
				FROM receipt_reviews rr
				INNER JOIN receipt_images ri ON ri.id = rr.receipt_image_id
				WHERE ri.user_id = $1
			)
	`, userID).Scan(&averageConfidence, &acceptedAccuracy); err != nil {
		return nil, err
	}
	summary.AverageConfidence = nullableFloat(averageConfidence)
	summary.AcceptedAccuracyRate = nullableFloat(acceptedAccuracy)

	groupRows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, description, status, total_images, queued_images, processing_images,
		       completed_images, failed_images, reviewed_images, export_count, created_at, updated_at
		FROM receipt_groups
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 5
	`, userID)
	if err != nil {
		return nil, err
	}
	defer groupRows.Close()

	for groupRows.Next() {
		var group domain.ReceiptGroup
		var description pgtype.Text
		if err := groupRows.Scan(
			&group.ID,
			&group.UserID,
			&group.Name,
			&description,
			&group.Status,
			&group.TotalImages,
			&group.QueuedImages,
			&group.ProcessingImages,
			&group.CompletedImages,
			&group.FailedImages,
			&group.ReviewedImages,
			&group.ExportCount,
			&group.CreatedAt,
			&group.UpdatedAt,
		); err != nil {
			return nil, err
		}
		group.Description = nullableText(description)
		summary.RecentGroups = append(summary.RecentGroups, group)
	}

	imageRows, err := r.db.Query(ctx, `
		SELECT id, group_id, user_id, original_filename, mime_type, file_size_bytes, checksum_sha256,
		       storage_bucket, storage_object_key, storage_url, upload_status, ocr_status, review_status,
		       ocr_attempt_count, last_error_code, last_error_message, receipt_type, overall_confidence,
		       processed_at, created_at, updated_at
		FROM receipt_images
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`, userID)
	if err != nil {
		return nil, err
	}
	defer imageRows.Close()

	for imageRows.Next() {
		image, err := scanReceiptImage(imageRows)
		if err != nil {
			return nil, err
		}
		summary.RecentImages = append(summary.RecentImages, *image)
	}

	return summary, nil
}
