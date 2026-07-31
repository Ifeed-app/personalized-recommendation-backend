UPDATE content_service.videos
SET status = LOWER(status);

ALTER TABLE content_service.videos
    ALTER COLUMN status SET DEFAULT 'draft';