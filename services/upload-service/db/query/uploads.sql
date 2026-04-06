-- name: CreateUpload :one
INSERT INTO upload_service.uploads (
    id,
    user_id,
    upload_type,
    status
) VALUES (
             $1, $2, $3, $4
         )
RETURNING id, user_id, upload_type, status, created_at, updated_at;

-- name: GetUploadByID :one
SELECT id, user_id, upload_type, status, created_at, updated_at
FROM upload_service.uploads
WHERE id = $1;

-- name: ListUploadsByUserID :many
SELECT id, user_id, upload_type, status, created_at, updated_at
FROM upload_service.uploads
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateUploadStatus :one
UPDATE upload_service.uploads
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, upload_type, status, created_at, updated_at;

-- name: DeleteUpload :exec
DELETE FROM upload_service.uploads
WHERE id = $1;