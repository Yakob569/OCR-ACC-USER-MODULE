# UI Design Prompt (Use With Stitch / Claude / Design AI)

## Context

We already have an existing product being built. Create modern, production-ready UI designs that *match the existing design system*.

This app is an **OCR Ledger** workflow:
- User logs in
- User creates a receipt group (batch)
- User uploads multiple receipt images
- Backend processes OCR asynchronously
- User views progress, history, results, reviews/corrections, retries, and exports to CSV

## Hard Constraints

- **Do not introduce a new color palette.** Reuse existing app theme/tokens/colors/typography components.
- **Do not say “use purple / use this color”.** Assume the app already defines tokens (CSS variables or a design system).
- Make it feel **very modern**, clean, high-contrast, and highly usable.
- Responsive: **mobile + desktop** layouts must both be designed.
- Accessibility: keyboard navigation, focus states, readable contrast, empty/loading/error states.

## Deliverables

We already have the unauthenticated pages:
- Landing
- Create account (register)
- Login

Create only the **post-login experience**:
1. Authenticated app shell layout (post-login)
2. Dashboard / Home page (post-login)
3. History / Groups list page
4. Group detail page (progress + images + results + exports)
5. Image detail page (OCR result + review + retry)
6. Export page (CSV export creation + export history)

Provide:
- page layouts (wireframe-level structure + final UI)
- key components + states
- suggested microcopy (button labels, empty states, errors)

## Information Architecture (Routes)

- `/app` App shell (default redirects to Dashboard)
- `/app/dashboard` Dashboard
- `/app/groups` History (groups)
- `/app/groups/:groupId` Group detail
- `/app/images/:imageId` Image detail
- `/app/exports` Exports (optional: can also be inside group detail)
- `/app/settings` (minimal placeholder is fine)

## Data Model (What The UI Shows)

### Dashboard Summary
- Total groups
- Total scans
- Successful scans
- Failed scans
- Needs-review scans
- Average confidence (optional)
- Accepted accuracy rate (optional)
- Recent groups list
- Recent images list

### Receipt Group
- name, description
- status: `draft | uploading | queued | processing | completed | completed_with_failures | failed | archived`
- counters: total, queued, processing, completed, failed, reviewed, export_count
- created_at, updated_at

### Receipt Image
- original_filename, mime_type, file_size_bytes
- upload_status: `pending | uploaded | upload_failed`
- ocr_status: `queued | processing | completed | failed | needs_review`
- review_status: `pending | reviewed | accepted | rejected`
- attempt count, last error (code/message)
- receipt type, confidence

### OCR Result + Review
- extracted fields + items (structured)
- warnings
- corrected fields (JSON)
- review notes

### Export
- format (CSV)
- selected columns
- row count
- created_at

## App Shell (Post-Login)

Design an app shell with:
- Left navigation (desktop) that collapses to bottom nav or drawer (mobile)
- Top bar with:
  - page title
  - global search (optional)
  - user menu (profile/logout)
  - “New Group” primary action (or prominent placement)
- Main content area with good spacing and a clear content grid

Navigation items:
- Dashboard
- History (Groups)
- Exports
- Settings

## Page Specs (Post-Login Only)

### 1) Dashboard (`/app/dashboard`)

Above-the-fold:
- Summary cards (Total scans, Successful, Failed, Needs review)
- A compact “Activity” strip (last 7 days) is optional if it fits the existing system

Main content:
- Recent groups table/list with:
  - group name, status pill, counters, updated time
  - row click goes to group detail
- Recent images list:
  - filename, status, confidence, last updated
  - click goes to image detail

Empty state (new account):
- “Create your first group” primary action
- short explanation of the flow

### 2) History / Groups (`/app/groups`)

Primary action:
- “New Group”

Filters:
- Status filter
- Search by name
- Sort by updated_at / created_at

Group list presentation:
- Desktop: table
- Mobile: cards

Row contents:
- Name + description
- Status pill
- Progress: (completed/total) and small segmented bar (queued/processing/completed/failed)
- Updated timestamp

### 3) Create Group (Modal or Page)

Inputs:
- name (required)
- description (optional)

After create:
- route to group detail

### 4) Group Detail (`/app/groups/:groupId`)

Header:
- Group name, description, status pill
- Progress summary and counters
- Primary CTA: “Upload receipts”
- Secondary actions: “Create CSV export”, “Refresh” (optional)

Sections:
1. Progress panel
  - segmented progress bar
  - counts: queued, processing, completed, failed, needs review
2. Upload area
  - drag & drop zone
  - multi-file picker
  - file constraints note
  - show upload in-flight list (filename + size + status)
3. Images table/list
  - filename
  - upload_status
  - ocr_status
  - review_status
  - confidence
  - actions: “View”, “Retry” (only if failed), “Review” (if needs_review)
4. Results preview (optional)
  - show top extracted fields for completed images (e.g., merchant, date, total)
5. Exports panel
  - list export history (created_at, row_count, status if needed)
  - “Create export” button

Empty states:
- No images uploaded yet
- No results yet
- No exports yet

### 5) Image Detail (`/app/images/:imageId`)

Layout:
- Two-column on desktop, stacked on mobile

Left column:
- Image preview (if available) or placeholder
- Metadata (filename, size, mime, attempt count)
- Status timeline (upload -> queued -> processing -> completed/failed -> review)

Right column:
- OCR result viewer:
  - Key fields section (merchant, date, total, tax, currency, etc. as generic placeholders)
  - Line items table (if present)
  - Warnings (if present)
- Review section:
  - Toggle accept/reject
  - Quality label dropdown (e.g., good/ok/bad)
  - Corrected fields editor: show structured form if possible, otherwise JSON editor in “advanced”
  - Review notes textarea
  - Submit review button
- Retry section:
  - show only if failed or needs_review and retry is allowed
  - retry button with confirmation

Failure details:
- show last error code/message in a small “Details” panel

### 6) Exports (`/app/exports` or within group)

Create export:
- Select group (if global exports page)
- Multi-select columns
- Toggle “Include corrected values”
- Preview row count (optional placeholder)
- Create button

Export history:
- list/table with created_at, group, row_count, format
- download link/button (if available) or “copy link” placeholder

## Component Guidance

Design these reusable components (matching existing system):
- Status pills for group/image statuses (use tokens, do not define new colors)
- Progress segmented bar for group counters
- Tables that collapse cleanly to cards on mobile
- Empty state component with icon placeholder + CTA
- Toast notifications (success/failure)
- Error banner component
- Loading skeletons for lists and detail pages
- Confirmation dialog for retry

## Interaction & UX Notes

- Make “New Group” and “Upload receipts” obvious and fast.
- Prefer optimistic UI for uploads: show each file as soon as selected, then update status.
- Provide clear status language:
  - `queued`, `processing`, `needs review`, `completed`, `failed`
- Avoid clutter: keep advanced info (raw JSON, debug) behind an “Advanced” accordion.

## Output Format Requested From The Design AI

- Provide each page as a complete design with:
  - desktop layout
  - mobile layout
  - key states (empty/loading/error)
- Keep styling consistent with an existing app:
  - reuse existing tokens and spacing scale
  - do not propose a new color palette
