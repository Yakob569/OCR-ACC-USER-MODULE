# Step 04: OCR Repository Adapters

## Completed

- added Postgres repository adapters in `internal/adapters/repositories/ocr_repo_impl.go`
- implemented repository constructors for:
  - receipt groups
  - receipt images
  - OCR extractions
  - OCR jobs
- implemented CRUD-style access needed for the next application steps:
  - create/list/get groups
  - create/list/get images
  - update image status and OCR result state
  - upsert OCR extraction payloads
  - create/list/get/update OCR jobs
- added scanning helpers for nullable DB fields
- added compile-time interface assertions against the new core ports

## Covered From Main Plan

- persistence layer for group workflows
- persistence layer for uploaded image tracking
- persistence layer for OCR result storage
- persistence layer for DB-backed OCR jobs

## Remaining

- application services using these repositories
- group management endpoints
- multi-image upload endpoint
- MinIO integration
- OCR engine HTTP client integration
- DB-backed worker/job processing
- dashboard endpoints
- review endpoints
- CSV export endpoints

## Notes

- these repositories are intentionally low-level adapters only
- orchestration logic still belongs in services, not in the repository layer
