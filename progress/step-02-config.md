# Step 02: OCR Engine And MinIO Config Contract

## Completed

- extended `internal/config/config.go` with OCR engine settings
- extended `internal/config/config.go` with MinIO settings
- added config parsing helpers for integer and boolean environment variables
- added validation for OCR timeout, OCR concurrency, group file limits, and upload size limits
- added config tests for:
  - default JWT rejection
  - OCR/MinIO env loading
  - invalid OCR timeout rejection

## Covered From Main Plan

- config additions for OCR engine
- config additions for MinIO
- upload size and batch limit configuration
- concurrency contract configuration for OCR orchestration

## Remaining

- domain models and repository interfaces for OCR/group features
- group management endpoints
- multi-image upload endpoint
- MinIO integration
- OCR engine HTTP client integration
- DB-backed worker/job processing
- dashboard endpoints
- review endpoints
- CSV export endpoints

## Notes

- OCR engine config is added without forcing current auth-only flows to immediately use it
- the Go service now has a stable config surface for the next implementation steps
