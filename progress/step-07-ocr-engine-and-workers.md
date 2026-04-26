# Step 07: OCR Engine Client And Job Workers

## Completed

- added external OCR engine HTTP client in `internal/adapters/ocrclient/http_adapter.go`
- added MinIO object download support in `internal/adapters/storage/minio_adapter.go`
- added DB-backed queued job claiming in `internal/adapters/repositories/ocr_repo_impl.go`
- added OCR job processing service in `internal/core/services/ocr_job_service.go`
- wired OCR engine client and OCR job service into `cmd/api/main.go`
- started background OCR workers on service startup

## Covered From Main Plan

- Go backend calling the external Python OCR engine
- DB-backed job claiming and processing
- object storage fetch for OCR execution
- extraction persistence path
- queued upload records now have an execution path

## Remaining

- image/result listing endpoints
- dashboard endpoints
- review endpoints
- CSV export endpoints
- retry endpoints
- stronger group status aggregation for mixed outcomes

## Notes

- workers use bounded concurrency from `OCR_ENGINE_MAX_CONCURRENCY`
- current group status updates are functional but still coarse; richer aggregate recomputation can be improved in later steps
