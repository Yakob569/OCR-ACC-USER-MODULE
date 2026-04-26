# Step 08: Result Read APIs

## Completed

- added `ReceiptQueryService` contract and implementation
- added extraction listing support in the OCR extraction repository
- extended `GroupHandler` with read endpoints for images and OCR results
- added protected endpoints:
  - `GET /api/v1/groups/{group_id}/images`
  - `GET /api/v1/groups/{group_id}/results`
  - `GET /api/v1/images/{image_id}`
  - `GET /api/v1/images/{image_id}/result`
- wired query service into `cmd/api/main.go`

## Covered From Main Plan

- frontend-readable image list for a group
- frontend-readable OCR result list for a group
- single image detail retrieval
- single OCR extraction retrieval

## Remaining

- dashboard endpoints
- review endpoints
- CSV export endpoints
- retry endpoints
- improved aggregate group status recomputation

## Notes

- these endpoints expose the current persisted OCR state and do not trigger OCR execution
- the read path is user-scoped through the existing auth middleware
