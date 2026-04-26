# Step 06: Multi-Image Group Upload Intake

## Completed

- added `ReceiptUploadService` implementation in `internal/core/services/receipt_upload_service.go`
- extended the OCR ports for upload results and OCR job creation inputs
- updated image persistence to accept pre-generated image IDs for MinIO object naming
- added MinIO-backed object storage adapter in `internal/adapters/storage/minio_adapter.go`
- extended `GroupHandler` with authenticated multi-image upload support
- added protected endpoint:
  - `POST /api/v1/groups/{group_id}/images`
- upload flow now:
  - validates authenticated user
  - validates target group ownership
  - reads multiple multipart files from `files`
  - validates image content types and size
  - uploads originals to object storage
  - persists `receipt_images`
  - creates queued `ocr_jobs`
  - updates group counters/status

## Covered From Main Plan

- multi-image upload endpoint in the Go backend
- MinIO object registration owned by the Go backend
- DB-backed image intake records
- DB-backed OCR job creation per uploaded image

## Remaining

- OCR engine HTTP client integration
- DB-backed worker/job processing loop
- OCR result persistence flow
- image/result listing endpoints
- dashboard endpoints
- review endpoints
- CSV export endpoints

## Notes

- this step creates queued OCR jobs but does not process them yet
- the next logical step is the worker/client path that calls the external Python OCR engine
