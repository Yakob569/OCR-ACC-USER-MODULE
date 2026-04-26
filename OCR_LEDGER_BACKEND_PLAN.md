# OCR Ledger Backend Plan

## Scope

This document defines the backend plan for the ledger OCR product with the correct service ownership:

- the Go service in `OCR-ACC-USER-MODULE` is the main backend
- the web app and mobile app call the Go service
- the Python OCR service is a separate deployed OCR engine
- the Python OCR service should remain stateless and should not own product data
- the Go service owns auth-aware user workflows, groups, image records, MinIO integration, OCR orchestration, analytics, and exports

This is the architecture we should implement going forward.

## Core Ownership Model

## Go Service Responsibilities

The Go backend should own:

- authenticated API endpoints
- user scoping
- group creation and management
- multi-image upload intake
- MinIO upload and object tracking
- DB schema and persistence
- OCR job orchestration
- calling the Python OCR service
- storing OCR results
- dashboard metrics
- review/correction state
- CSV export preparation

## Python OCR Service Responsibilities

The Python OCR service should only own:

- image preprocessing
- OCR extraction
- template parsing
- confidence scoring
- structured OCR response

It should not own:

- database access
- user state
- group state
- MinIO object registration
- dashboard metrics
- export history
- application-level auth

That separation keeps the OCR engine simple and lets the Go service remain the product backend.

## Authentication Model

Only authenticated users should be able to use OCR workflows.

The Go backend should treat the caller as an authenticated user because:

- users are registered through the auth service
- protected endpoints already use token validation
- the Go service can resolve the requesting `user_id` from the bearer token

So every OCR-related record should be scoped to the authenticated user:

- groups
- uploaded images
- OCR jobs
- review records
- exports

The frontend and mobile app should never directly call the Python OCR service.

## High-Level Request Flow

The correct request flow should be:

1. user signs in
2. frontend/mobile receives auth token
3. frontend/mobile calls the Go backend with bearer token
4. Go backend validates the token and resolves `user_id`
5. user creates a group such as `cafe-23`
6. user uploads one or many images to the Go backend
7. Go backend stores original images in MinIO
8. Go backend creates DB records and OCR jobs
9. Go backend asynchronously calls the Python OCR engine per image
10. Go backend stores OCR results and updates group status
11. frontend/mobile polls or fetches group progress/results from the Go backend

## Main Architectural Decision

The Go service is the orchestrator.

The Python OCR service is a worker-like extraction dependency exposed over HTTP.

This means the product is not “a Python OCR app with extra features.” It is:

- a Go product backend
- backed by a Python OCR microservice

## Config Requirements

The Go service should add config for the OCR engine base URL, for example:

- `OCR_ENGINE_BASE_URL`
- `OCR_ENGINE_TIMEOUT_SECONDS`
- `OCR_ENGINE_MAX_CONCURRENCY`

The Go service should also hold MinIO config and database config.

Example additional config direction:

- `JWT_SECRET`
- `DATABASE_URL`
- `MINIO_ACCESS_KEY_ID`
- `MINIO_SECRET_ACCESS_KEY`
- `MINIO_BUCKET_NAME`
- `MINIO_END_POINT`
- `MINIO_USE_SSL`
- `MINIO_CLAMAV_URL`
- `OCR_ENGINE_BASE_URL`
- `OCR_ENGINE_TIMEOUT_SECONDS`
- `OCR_ENGINE_MAX_CONCURRENCY`
- `OCR_GROUP_MAX_FILES`
- `OCR_MAX_FILE_SIZE_MB`

## OCR Engine Contract

The Python service currently exposes:

- `POST /api/v1/ocr/extract`

and returns a structured OCR response with:

- `success`
- `receipt_type`
- `fields`
- `items`
- `warnings`
- optional `raw_text`
- optional `debug`

The Go service should call this endpoint internally for each uploaded image when it is time to process it.

## Main Product Workflow

The intended product workflow should be:

1. authenticated user creates a group
2. authenticated user uploads many receipt images into that group
3. Go backend uploads the originals to MinIO and creates DB records immediately
4. Go backend fans out OCR processing tasks per image
5. each OCR task calls the Python OCR service
6. Go backend stores the OCR result for each image
7. Go backend aggregates group progress and summary metrics
8. user views group history, recent scans, status counts, and result details
9. user optionally reviews/corrects OCR results
10. user exports selected fields into CSV

## Batch Upload Handling

One major requirement is that the Go backend should accept multiple images in a single request.

That endpoint should support a user sending:

- `3` images
- `20` images
- `30` images

without the frontend needing to manage per-image OCR orchestration manually.

## Recommended Intake Pattern

The upload endpoint should:

1. accept an array/list of image files
2. validate file count, MIME type, and size
3. upload each original image to MinIO
4. create one DB record per image
5. create one OCR job per image
6. return quickly with accepted job metadata

The endpoint should not synchronously wait for all OCR work to finish if batches are large.

## Concurrency Model

Your idea of processing images concurrently from the Go service is correct, but it needs controlled concurrency.

We should not spawn unbounded goroutines and immediately fire `30` HTTP OCR requests without limits.

The better design is:

- create one logical job per image
- process jobs concurrently with a bounded worker pool
- each worker uses goroutines internally but respects a concurrency limit

## Recommended Pattern

Use bounded fan-out/fan-in.

### Fan-out

For each uploaded image:

- upload to MinIO
- create image record
- create OCR job
- dispatch work

### Processing

For each OCR job:

- read image metadata from DB
- fetch original bytes if needed from MinIO or use request-time bytes if still available
- call Python OCR endpoint
- parse/store the OCR response
- update status fields

### Fan-in

As jobs finish:

- update image status
- increment group counters
- compute aggregate group status
- expose progress to the frontend/mobile app

## Goroutines Guidance

Yes, goroutines are appropriate in the Go service, but use them in a disciplined way.

Recommended options:

### Option 1: In-process worker pool

Good for first implementation.

Flow:

- upload request creates DB records
- app pushes image jobs into an internal buffered work queue
- a fixed number of workers pull jobs
- each worker calls the OCR engine HTTP endpoint

Pros:

- simple to build first
- fast iteration

Cons:

- jobs live inside service memory unless also persisted
- restarts need recovery logic

### Option 2: DB-backed job polling workers

Better for reliability.

Flow:

- upload request persists jobs in DB with `queued` status
- background worker loop in Go polls queued jobs
- workers mark jobs `processing`
- workers call OCR engine
- workers mark jobs `completed` or `failed`

Pros:

- restart-safe
- easier retries
- better audit trail

Cons:

- slightly more implementation work

## Recommendation

For this product, use DB-backed jobs plus a bounded worker pool in Go.

That gives:

- persistence
- retries
- observable progress
- better behavior on restarts

## Why The Python Service Should Stay Stateless

The Python OCR service will likely be deployed independently on Render.

That means it is best used as:

- stateless HTTP processor
- horizontally replaceable
- simple to redeploy

If the Python service also owns DB state, MinIO state, group state, and analytics, the architecture becomes harder to maintain and scale.

## Endpoint Plan In The Go Service

The Go service should expose the product endpoints.

## Existing Auth/User Context

This repo already has auth patterns such as:

- protected routes
- token validation
- user context resolution

OCR endpoints should reuse that model so every request is user-bound.

## Group Endpoints

### `POST /api/v1/groups`

Create a user group.

Request:

- `name`
- `description` optional

Response:

- `id`
- `name`
- `status`
- `created_at`

### `GET /api/v1/groups`

List groups for the authenticated user.

### `GET /api/v1/groups/{group_id}`

Return group detail and counters.

### `PATCH /api/v1/groups/{group_id}`

Rename group or update description.

### `DELETE /api/v1/groups/{group_id}`

Archive or soft delete a group.

## Upload Endpoints

### `POST /api/v1/groups/{group_id}/images`

Accept multiple image files in one request.

Recommended request shape:

- multipart form-data
- field name like `files`
- repeatable multiple file parts

Optional metadata:

- `source`
- `notes`

Response should include:

- group id
- accepted files count
- rejected files count
- created image ids
- created job ids
- immediate group status

### `GET /api/v1/groups/{group_id}/images`

List all images in the group with statuses.

### `GET /api/v1/images/{image_id}`

Return detailed image record including OCR status and storage metadata.

## Result Endpoints

### `GET /api/v1/images/{image_id}/result`

Return OCR extraction result for a single image.

### `GET /api/v1/groups/{group_id}/results`

Return paginated OCR results for a whole group.

## Dashboard Endpoints

### `GET /api/v1/dashboard/summary`

Return:

- total scans
- successful scans
- failed scans
- needs review count
- accepted accuracy rate
- average confidence
- recent groups
- recent processed images

### `GET /api/v1/dashboard/recent`

Return recent activity for the user.

## Review Endpoints

### `POST /api/v1/images/{image_id}/review`

Store user review and correction.

### `GET /api/v1/groups/{group_id}/review-summary`

Return group review metrics.

## Export Endpoints

### `POST /api/v1/groups/{group_id}/exports/csv`

Generate CSV with selected columns.

### `GET /api/v1/groups/{group_id}/exports`

List export history.

## Retry/Operations Endpoints

### `POST /api/v1/images/{image_id}/retry`

Retry one failed OCR image.

### `POST /api/v1/groups/{group_id}/retry-failures`

Retry all failed images in a group.

### `GET /api/v1/jobs/{job_id}`

Return OCR job status.

## Database Schema Direction

The database should live with the Go service, not the Python OCR service.

## Main Tables

### `receipt_groups`

Suggested fields:

- `id UUID PK`
- `user_id UUID`
- `name TEXT`
- `description TEXT NULL`
- `status TEXT`
- `total_images INT DEFAULT 0`
- `queued_images INT DEFAULT 0`
- `processing_images INT DEFAULT 0`
- `completed_images INT DEFAULT 0`
- `failed_images INT DEFAULT 0`
- `reviewed_images INT DEFAULT 0`
- `export_count INT DEFAULT 0`
- `created_at`
- `updated_at`

### `receipt_images`

Suggested fields:

- `id UUID PK`
- `group_id UUID`
- `user_id UUID`
- `original_filename TEXT`
- `mime_type TEXT`
- `file_size_bytes BIGINT`
- `checksum_sha256 TEXT`
- `storage_bucket TEXT`
- `storage_object_key TEXT`
- `storage_url TEXT NULL`
- `upload_status TEXT`
- `ocr_status TEXT`
- `review_status TEXT`
- `ocr_attempt_count INT DEFAULT 0`
- `last_error_code TEXT NULL`
- `last_error_message TEXT NULL`
- `receipt_type TEXT NULL`
- `overall_confidence NUMERIC(5,4) NULL`
- `processed_at TIMESTAMPTZ NULL`
- `created_at`
- `updated_at`

### `receipt_extractions`

Suggested fields:

- `id UUID PK`
- `receipt_image_id UUID UNIQUE`
- `success BOOLEAN`
- `receipt_type TEXT`
- `fields_json JSONB`
- `items_json JSONB`
- `warnings_json JSONB`
- `raw_text TEXT NULL`
- `debug_json JSONB NULL`
- `ocr_engine_url TEXT NULL`
- `ocr_engine_version TEXT NULL`
- `pipeline_version TEXT NULL`
- `created_at`
- `updated_at`

### `receipt_reviews`

Suggested fields:

- `id UUID PK`
- `receipt_image_id UUID`
- `reviewed_by_user_id UUID`
- `quality_label TEXT`
- `is_accepted BOOLEAN`
- `corrected_fields_json JSONB`
- `review_notes TEXT NULL`
- `reviewed_at TIMESTAMPTZ`
- `created_at`

### `group_exports`

Suggested fields:

- `id UUID PK`
- `group_id UUID`
- `exported_by_user_id UUID`
- `format TEXT`
- `selected_columns_json JSONB`
- `row_count INT`
- `storage_bucket TEXT NULL`
- `storage_object_key TEXT NULL`
- `storage_url TEXT NULL`
- `created_at`

### `ocr_jobs`

Suggested fields:

- `id UUID PK`
- `receipt_image_id UUID`
- `group_id UUID`
- `user_id UUID`
- `status TEXT`
- `attempt_count INT DEFAULT 0`
- `max_attempts INT DEFAULT 3`
- `queued_at TIMESTAMPTZ`
- `started_at TIMESTAMPTZ NULL`
- `finished_at TIMESTAMPTZ NULL`
- `worker_id TEXT NULL`
- `error_code TEXT NULL`
- `error_message TEXT NULL`
- `created_at`
- `updated_at`

## MinIO Design

The Go service should upload and register original images in MinIO.

The Python OCR service should not own object storage metadata.

## Recommended Stored Fields

For each uploaded image, persist:

- MinIO bucket
- object key
- file size
- checksum
- MIME type
- optional URL

## Recommended Object Key Pattern

Original uploads:

`receipts/{user_id}/{group_id}/original/{image_id}-{safe_filename}`

Optional future processed artifacts:

`receipts/{user_id}/{group_id}/processed/{image_id}-{variant}.png`

CSV exports:

`exports/{user_id}/{group_id}/{export_id}.csv`

## OCR Call Strategy

The Go service should call the Python OCR engine over HTTP.

Recommended per-image flow:

1. read image bytes
2. build multipart request
3. `POST {OCR_ENGINE_BASE_URL}/api/v1/ocr/extract`
4. parse JSON response
5. persist normalized result in DB
6. update image and job statuses

## Timeouts And Reliability

Each OCR call should have:

- request timeout
- retry policy for transient failures
- circuit-breaker style protection later if needed

Do not let stuck OCR calls block the entire worker pool forever.

## Group Progress Logic

Group status should be derived from child image/job states:

- all queued: `queued`
- any processing: `processing`
- all completed, none failed: `completed`
- some completed, some failed: `completed_with_failures`
- all failed: `failed`

The frontend can poll group detail and group images endpoints to render progress bars and recent results.

## Dashboard Metrics

The home screen should show both machine and human metrics.

## Suggested Metrics

- `total_scans`
- `successful_scans`
- `failed_scans`
- `needs_review_scans`
- `average_confidence`
- `accepted_accuracy_rate`
- `recent_groups`
- `recent_processed_images`

## Important Metric Definitions

Keep these distinct:

- processing success:
  - OCR job completed with usable extracted data
- reviewed accuracy:
  - user explicitly marked result as accurate

Those are different and both matter.

## CSV Export Direction

CSV export belongs to the Go backend because it has:

- group metadata
- image metadata
- OCR result data
- review corrections
- user field selection

Users should be able to choose columns like:

- group name
- source filename
- merchant name
- invoice number
- receipt number
- date
- customer name
- subtotal
- tax
- total
- confidence
- warning count

Version 1 should export one row per receipt image.

## Suggested Implementation Order

Build in this order:

1. add DB schema in the Go service
2. add config for OCR engine and MinIO
3. add authenticated group endpoints
4. add multi-image upload endpoint
5. add MinIO upload logic
6. add OCR job table and worker processing
7. add HTTP client integration to Python OCR engine
8. persist OCR results
9. add dashboard endpoints
10. add review endpoints
11. add CSV export endpoints

## Final Recommendation

The correct model is:

- Go service is the main backend product API
- authenticated users call the Go service only
- Go service stores data, owns workflows, and manages concurrency
- Python OCR service is an external stateless OCR processor

For batches of `20-30` images, the Go backend should accept them in one request, persist them, and process them concurrently through bounded goroutine-backed workers that call the Python OCR service.
