-- OCR ledger product schema owned by the Go backend.
-- The Python OCR service remains stateless and only performs extraction.

CREATE TABLE IF NOT EXISTS receipt_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (
        status IN (
            'draft',
            'uploading',
            'queued',
            'processing',
            'completed',
            'completed_with_failures',
            'failed',
            'archived'
        )
    ),
    total_images INT NOT NULL DEFAULT 0 CHECK (total_images >= 0),
    queued_images INT NOT NULL DEFAULT 0 CHECK (queued_images >= 0),
    processing_images INT NOT NULL DEFAULT 0 CHECK (processing_images >= 0),
    completed_images INT NOT NULL DEFAULT 0 CHECK (completed_images >= 0),
    failed_images INT NOT NULL DEFAULT 0 CHECK (failed_images >= 0),
    reviewed_images INT NOT NULL DEFAULT 0 CHECK (reviewed_images >= 0),
    export_count INT NOT NULL DEFAULT 0 CHECK (export_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, name)
);

CREATE TABLE IF NOT EXISTS receipt_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES receipt_groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_filename TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    file_size_bytes BIGINT NOT NULL CHECK (file_size_bytes >= 0),
    checksum_sha256 TEXT NOT NULL,
    storage_bucket TEXT NOT NULL,
    storage_object_key TEXT NOT NULL,
    storage_url TEXT,
    upload_status TEXT NOT NULL DEFAULT 'pending' CHECK (
        upload_status IN ('pending', 'uploaded', 'upload_failed')
    ),
    ocr_status TEXT NOT NULL DEFAULT 'queued' CHECK (
        ocr_status IN ('queued', 'processing', 'completed', 'failed', 'needs_review')
    ),
    review_status TEXT NOT NULL DEFAULT 'pending' CHECK (
        review_status IN ('pending', 'reviewed', 'accepted', 'rejected')
    ),
    ocr_attempt_count INT NOT NULL DEFAULT 0 CHECK (ocr_attempt_count >= 0),
    last_error_code TEXT,
    last_error_message TEXT,
    receipt_type TEXT,
    overall_confidence NUMERIC(5,4) CHECK (
        overall_confidence IS NULL OR (overall_confidence >= 0 AND overall_confidence <= 1)
    ),
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS receipt_extractions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    receipt_image_id UUID NOT NULL UNIQUE REFERENCES receipt_images(id) ON DELETE CASCADE,
    success BOOLEAN NOT NULL,
    receipt_type TEXT,
    fields_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    items_json JSONB NOT NULL DEFAULT '[]'::JSONB,
    warnings_json JSONB NOT NULL DEFAULT '[]'::JSONB,
    raw_text TEXT,
    debug_json JSONB,
    ocr_engine_url TEXT,
    ocr_engine_version TEXT,
    pipeline_version TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS receipt_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    receipt_image_id UUID NOT NULL REFERENCES receipt_images(id) ON DELETE CASCADE,
    reviewed_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quality_label TEXT NOT NULL CHECK (
        quality_label IN ('accurate', 'partially_accurate', 'inaccurate')
    ),
    is_accepted BOOLEAN NOT NULL,
    corrected_fields_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    review_notes TEXT,
    reviewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS group_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES receipt_groups(id) ON DELETE CASCADE,
    exported_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    format TEXT NOT NULL CHECK (format IN ('csv')),
    selected_columns_json JSONB NOT NULL DEFAULT '[]'::JSONB,
    row_count INT NOT NULL DEFAULT 0 CHECK (row_count >= 0),
    storage_bucket TEXT,
    storage_object_key TEXT,
    storage_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ocr_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    receipt_image_id UUID NOT NULL REFERENCES receipt_images(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES receipt_groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'queued' CHECK (
        status IN ('queued', 'processing', 'completed', 'failed', 'retrying', 'cancelled')
    ),
    attempt_count INT NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    max_attempts INT NOT NULL DEFAULT 3 CHECK (max_attempts > 0),
    queued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    worker_id TEXT,
    error_code TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_receipt_groups_user_id
ON receipt_groups(user_id);

CREATE INDEX IF NOT EXISTS idx_receipt_groups_status
ON receipt_groups(status);

CREATE INDEX IF NOT EXISTS idx_receipt_images_group_id
ON receipt_images(group_id);

CREATE INDEX IF NOT EXISTS idx_receipt_images_user_id
ON receipt_images(user_id);

CREATE INDEX IF NOT EXISTS idx_receipt_images_ocr_status
ON receipt_images(ocr_status);

CREATE INDEX IF NOT EXISTS idx_receipt_images_checksum_sha256
ON receipt_images(checksum_sha256);

CREATE UNIQUE INDEX IF NOT EXISTS idx_receipt_images_group_storage_object_key
ON receipt_images(group_id, storage_object_key);

CREATE INDEX IF NOT EXISTS idx_receipt_reviews_receipt_image_id
ON receipt_reviews(receipt_image_id);

CREATE INDEX IF NOT EXISTS idx_group_exports_group_id
ON group_exports(group_id);

CREATE INDEX IF NOT EXISTS idx_ocr_jobs_status
ON ocr_jobs(status);

CREATE INDEX IF NOT EXISTS idx_ocr_jobs_group_id
ON ocr_jobs(group_id);

CREATE INDEX IF NOT EXISTS idx_ocr_jobs_user_id
ON ocr_jobs(user_id);

CREATE INDEX IF NOT EXISTS idx_ocr_jobs_receipt_image_id
ON ocr_jobs(receipt_image_id);

DROP TRIGGER IF EXISTS trg_receipt_groups_updated_at ON receipt_groups;
CREATE TRIGGER trg_receipt_groups_updated_at
    BEFORE UPDATE ON receipt_groups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS trg_receipt_images_updated_at ON receipt_images;
CREATE TRIGGER trg_receipt_images_updated_at
    BEFORE UPDATE ON receipt_images
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS trg_receipt_extractions_updated_at ON receipt_extractions;
CREATE TRIGGER trg_receipt_extractions_updated_at
    BEFORE UPDATE ON receipt_extractions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS trg_ocr_jobs_updated_at ON ocr_jobs;
CREATE TRIGGER trg_ocr_jobs_updated_at
    BEFORE UPDATE ON ocr_jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
