# Step 01: OCR Schema Foundation

## Completed

- added the main backend planning document in `OCR_LEDGER_BACKEND_PLAN.md`
- added the first OCR ledger migration in `db/migrations/20260426130000_add_ocr_ledger_schema.sql`
- created the core product tables described in the main plan:
  - `receipt_groups`
  - `receipt_images`
  - `receipt_extractions`
  - `receipt_reviews`
  - `group_exports`
  - `ocr_jobs`
- added indexes, status constraints, and `updated_at` triggers for the mutable tables
- kept ownership in the Go backend, consistent with the main plan

## Covered From Main Plan

- DB schema for groups
- DB schema for uploaded images
- DB schema for OCR result persistence
- DB schema for review/correction tracking
- DB schema for export history
- DB schema for OCR job orchestration

## Remaining

- config additions for OCR engine and MinIO
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

- the Python OCR service still remains external and stateless
- the Go service is still the orchestrator and source of truth for product data
