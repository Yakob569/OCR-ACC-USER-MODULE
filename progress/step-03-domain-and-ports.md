# Step 03: OCR Domain And Ports

## Completed

- added OCR product domain models in `internal/core/domain/ocr.go`
- added status constants for groups, uploads, OCR, reviews, and jobs
- added core structs for:
  - receipt groups
  - receipt images
  - OCR extractions
  - OCR jobs
  - dashboard summary
- added service and repository contracts in `internal/core/ports/ocr_ports.go`
- added contracts for:
  - group persistence
  - image persistence
  - extraction persistence
  - OCR job persistence
  - object storage
  - OCR engine client
  - group service
  - upload service
  - job processing service
  - dashboard service

## Covered From Main Plan

- core product backend domain layer
- contracts for DB-backed job orchestration
- contracts for MinIO integration
- contracts for external Python OCR engine calls
- contracts for group/image workflows

## Remaining

- repository implementations for the new OCR tables
- group management endpoints
- multi-image upload endpoint
- MinIO integration
- OCR engine HTTP client integration
- DB-backed worker/job processing
- dashboard endpoints
- review endpoints
- CSV export endpoints

## Notes

- this step intentionally stops at contracts so upcoming adapters can be implemented cleanly
- transport and persistence details remain outside the core layer
