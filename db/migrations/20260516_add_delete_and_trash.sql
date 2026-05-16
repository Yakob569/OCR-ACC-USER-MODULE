-- Migration to add soft-delete for groups and trash table for images
BEGIN;

-- Add deleted_at to receipt_groups for soft-delete
ALTER TABLE receipt_groups ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- Create receipt_images_trash table
CREATE TABLE receipt_images_trash (
    id UUID PRIMARY KEY,
    group_id UUID NOT NULL,
    original_filename TEXT NOT NULL,
    storage_url TEXT NOT NULL,
    ocr_status TEXT NOT NULL,
    review_status TEXT NOT NULL,
    ocr_attempt_count INT DEFAULT 0,
    overall_confidence FLOAT,
    last_error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    trashed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMIT;
