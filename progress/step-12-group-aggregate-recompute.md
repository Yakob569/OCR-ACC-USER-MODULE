# Step 12: Group Aggregate Recompute

## Completed

- added `RefreshAggregateState` to `ReceiptGroupRepository`
- implemented full aggregate recomputation in the repository from:
  - `receipt_images`
  - `group_exports`
- group status is now recomputed from persisted data instead of relying only on incremental counters
- wired recomputation into:
  - upload intake
  - OCR job processing
  - OCR failure handling
  - review submission
  - retry flow
  - export creation

## Covered From Main Plan

- stronger aggregate group status recomputation
- safer counter/status consistency across retries, reviews, exports, and OCR processing

## Remaining

- no major planned product slice remains from the current implementation plan

## Notes

- aggregate state now comes from stored truth rather than only step-by-step counter mutations
- this reduces drift risk when multiple flows touch the same group over time
