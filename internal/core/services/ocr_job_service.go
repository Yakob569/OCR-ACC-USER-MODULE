package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cashflow/auth-service/internal/core/domain"
	"github.com/cashflow/auth-service/internal/core/ports"
	"github.com/google/uuid"
)

type ocrJobService struct {
	jobRepo        ports.OCRJobRepository
	imageRepo      ports.ReceiptImageRepository
	extractionRepo ports.OCRExtractionRepository
	groupRepo      ports.ReceiptGroupRepository
	objectStorage  ports.ObjectStorageService
	ocrEngine      ports.OCREngineService
	workerID       string
	concurrency    int
}

func NewOCRJobService(
	jobRepo ports.OCRJobRepository,
	imageRepo ports.ReceiptImageRepository,
	extractionRepo ports.OCRExtractionRepository,
	groupRepo ports.ReceiptGroupRepository,
	objectStorage ports.ObjectStorageService,
	ocrEngine ports.OCREngineService,
	concurrency int,
) ports.OCRJobService {
	if concurrency <= 0 {
		concurrency = 1
	}

	return &ocrJobService{
		jobRepo:        jobRepo,
		imageRepo:      imageRepo,
		extractionRepo: extractionRepo,
		groupRepo:      groupRepo,
		objectStorage:  objectStorage,
		ocrEngine:      ocrEngine,
		workerID:       uuid.NewString(),
		concurrency:    concurrency,
	}
}

func (s *ocrJobService) StartWorkers(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	sem := make(chan struct{}, s.concurrency)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			available := s.concurrency - len(sem)
			if available <= 0 {
				continue
			}

			jobs, err := s.jobRepo.ClaimQueued(ctx, s.workerID, available)
			if err != nil {
				log.Printf("[OCRJobService] Failed to claim jobs: %v", err)
				continue
			}

			for _, job := range jobs {
				jobID := job.ID
				sem <- struct{}{}
				go func() {
					defer func() { <-sem }()
					if err := s.ProcessJob(ctx, jobID); err != nil {
						log.Printf("[OCRJobService] Failed to process job %s: %v", jobID, err)
					}
				}()
			}
		}
	}
}

func (s *ocrJobService) ProcessJob(ctx context.Context, jobID uuid.UUID) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	if err := s.jobRepo.IncrementAttempt(ctx, jobID); err != nil {
		return err
	}

	image, err := s.imageRepo.GetByID(ctx, job.ReceiptImageID)
	if err != nil {
		_ = s.failJob(ctx, job, "image_not_found", err.Error())
		return err
	}

	if err := s.imageRepo.UpdateStatuses(ctx, image.ID, image.UploadStatus, domain.OCRStatusProcessing, image.ReviewStatus); err != nil {
		return err
	}
	_ = s.groupRepo.IncrementImageCounters(ctx, image.GroupID, 0, -1, 1, 0, 0, 0, 0)
	_ = s.groupRepo.UpdateStatus(ctx, image.GroupID, domain.GroupStatusProcessing)

	imageBytes, err := s.objectStorage.DownloadReceiptImage(ctx, image.StorageBucket, image.StorageObjectKey)
	if err != nil {
		_ = s.failJob(ctx, job, "object_download_failed", err.Error())
		return err
	}

	result, err := s.ocrEngine.Extract(ctx, image.OriginalFilename, image.MIMEType, bytes.NewReader(imageBytes))
	if err != nil {
		_ = s.failJob(ctx, job, "ocr_engine_failed", err.Error())
		return err
	}

	result.Extraction.ReceiptImageID = image.ID
	engineURL := "external"
	result.Extraction.OCREngineURL = &engineURL

	if err := s.extractionRepo.Upsert(ctx, result.Extraction); err != nil {
		_ = s.failJob(ctx, job, "extraction_persist_failed", err.Error())
		return err
	}

	if err := s.imageRepo.UpdateProcessingResult(ctx, image.ID, result.OCRStatus, result.ReceiptType, result.Confidence, nil, nil); err != nil {
		_ = s.failJob(ctx, job, "image_update_failed", err.Error())
		return err
	}

	if err := s.jobRepo.UpdateStatus(ctx, job.ID, domain.JobStatusCompleted, &s.workerID, nil, nil); err != nil {
		return err
	}

	_ = s.groupRepo.IncrementImageCounters(ctx, image.GroupID, 0, 0, -1, 1, 0, 0, 0)
	_ = s.groupRepo.UpdateStatus(ctx, image.GroupID, domain.GroupStatusCompleted)
	return nil
}

func (s *ocrJobService) failJob(ctx context.Context, job *domain.OCRJob, code, message string) error {
	errorCode := code
	errorMessage := message

	if err := s.imageRepo.UpdateProcessingResult(ctx, job.ReceiptImageID, domain.OCRStatusFailed, nil, nil, &errorCode, &errorMessage); err != nil {
		return err
	}
	if err := s.jobRepo.UpdateStatus(ctx, job.ID, domain.JobStatusFailed, &s.workerID, &errorCode, &errorMessage); err != nil {
		return err
	}
	_ = s.groupRepo.IncrementImageCounters(ctx, job.GroupID, 0, 0, -1, 0, 1, 0, 0)
	_ = s.groupRepo.UpdateStatus(ctx, job.GroupID, domain.GroupStatusCompletedWithFailures)
	return fmt.Errorf("%s: %s", code, message)
}
