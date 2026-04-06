-- name: GetVideoByID :one
SELECT
    id,
    creator_id,
    title,
    description,
    storage_key,
    thumbnail_key,
    mime_type,
    duration_seconds,
    status,
    created_at,
    updated_at
FROM content_service.videos
WHERE id = $1;

-- name: CreateVideo :one
INSERT INTO content_service.videos (
    id,
    creator_id,
    title,
    description,
    storage_key,
    thumbnail_key,
    mime_type,
    duration_seconds,
    status
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8, $9
         )
RETURNING
    id,
    creator_id,
    title,
    description,
    storage_key,
    thumbnail_key,
    mime_type,
    duration_seconds,
    status,
    created_at,
    updated_at;

-- name: ListVideosByCreatorID :many
SELECT
    id,
    creator_id,
    title,
    description,
    storage_key,
    thumbnail_key,
    mime_type,
    duration_seconds,
    status,
    created_at,
    updated_at
FROM content_service.videos
WHERE creator_id = $1
ORDER BY created_at DESC;

-- name: ListPublishedVideos :many
SELECT
    id,
    creator_id,
    title,
    description,
    storage_key,
    thumbnail_key,
    mime_type,
    duration_seconds,
    status,
    created_at,
    updated_at
FROM content_service.videos
WHERE status = 'published'
ORDER BY created_at DESC;

-- name: UpdateVideoStatus :one
UPDATE content_service.videos
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    creator_id,
    title,
    description,
    storage_key,
    thumbnail_key,
    mime_type,
    duration_seconds,
    status,
    created_at,
    updated_at;

-- name: DeleteVideo :exec
DELETE FROM content_service.videos
WHERE id = $1;