# Step 09: Dashboard Summary API

## Completed

- added `DashboardRepository` contract
- added `DashboardService` implementation
- added `DashboardHandler`
- added protected endpoint:
  - `GET /api/v1/dashboard/summary`
- wired dashboard repository/service/handler into server startup
- dashboard summary now returns:
  - total groups
  - total scans
  - successful scans
  - failed scans
  - needs review scans
  - average confidence
  - accepted accuracy rate
  - recent groups
  - recent images

## Covered From Main Plan

- home screen summary metrics
- home screen recent groups
- home screen recent processed image records

## Remaining

- review endpoints
- retry endpoints
- CSV export endpoints
- stronger aggregate group status recomputation

## Notes

- accepted accuracy rate stays `null` until review data exists
- dashboard uses persisted data only and does not trigger OCR work
