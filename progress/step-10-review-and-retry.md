# Step 10: Review And Retry APIs

## Completed

- added `ReceiptReviewRepository` contract and implementation
- added `ReceiptReviewService`
- added `OCRRetryService`
- added request model for image review submission
- added protected endpoints:
  - `POST /api/v1/images/{image_id}/review`
  - `POST /api/v1/images/{image_id}/retry`
- wired review and retry services into server startup

## Covered From Main Plan

- human review submission
- corrected field persistence
- accepted/inaccurate feedback capture
- retry flow for failed or needs-review images

## Remaining

- CSV export endpoints
- stronger aggregate group status recomputation

## Notes

- retry creates a new queued OCR job and resets the image OCR status back to queued
- review increments group reviewed count the first time a review is created for an image
