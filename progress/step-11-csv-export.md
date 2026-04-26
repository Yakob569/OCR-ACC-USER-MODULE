# Step 11: CSV Export APIs

## Completed

- added `GroupExportRepository` contract and implementation
- added `GroupExportService`
- extended object storage with export upload support
- added request model for CSV export generation
- added protected endpoints:
  - `POST /api/v1/groups/{group_id}/exports/csv`
  - `GET /api/v1/groups/{group_id}/exports`
- wired export repository/service into server startup
- export flow now:
  - loads group images
  - loads OCR extraction data
  - optionally overlays corrected review values
  - generates one CSV row per receipt image
  - uploads the CSV artifact to object storage
  - persists export history in `group_exports`

## Covered From Main Plan

- configurable CSV export generation
- export history
- MinIO-backed CSV artifact storage

## Remaining

- stronger aggregate group status recomputation

## Notes

- selected columns can include metadata columns and extracted field keys
- corrected review values take precedence when `include_corrected_values` is true
