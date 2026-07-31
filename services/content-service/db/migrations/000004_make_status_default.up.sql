ALTER TABLE content_service.videos
    ALTER COLUMN status SET DEFAULT 'DRAFT';

UPDATE content_service.videos
SET status = UPPER(status);