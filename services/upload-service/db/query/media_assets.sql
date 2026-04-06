-- name: CreateMediaAsset :one
INSERT INTO upload_service.media_assets (
    id,
    upload_id,
    owner_id,
    asset_type,
    storage_key,
    file_name,
    mime_type,
    size_bytes,
    checksum,
    status
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
         )
RETURNING
    id,
    upload_id,
    owner_id,
    asset_type,
    storage_key,
    file_name,
    mime_type,
    size_bytes,
    checksum,
    status,
    created_at,
    updated_at;

-- name: GetMediaAssetByID :one
SELECT
    id,
    upload_id,
    owner_id,
    asset_type,
    storage_key,
    file_name,
    mime_type,
    size_bytes,
    checksum,
    status,
    created_at,
    updated_at
FROM upload_service.media_assets
WHERE id = $1;

-- name: GetMediaAssetByStorageKey :one
SELECT
    id,
    upload_id,
    owner_id,
    asset_type,
    storage_key,
    file_name,
    mime_type,
    size_bytes,
    checksum,
    status,
    created_at,
    updated_at
FROM upload_service.media_assets
WHERE storage_key = $1;

-- name: ListMediaAssetsByUploadID :many
SELECT
    id,
    upload_id,
    owner_id,
    asset_type,
    storage_key,
    file_name,
    mime_type,
    size_bytes,
    checksum,
    status,
    created_at,
    updated_at
FROM upload_service.media_assets
WHERE upload_id = $1
ORDER BY created_at ASC;

-- name: ListMediaAssetsByOwnerID :many
SELECT
    id,
    upload_id,
    owner_id,
    asset_type,
    storage_key,
    file_name,
    mime_type,
    size_bytes,
    checksum,
    status,
    created_at,
    updated_at
FROM upload_service.media_assets
WHERE owner_id = $1
ORDER BY created_at DESC;

-- name: UpdateMediaAssetStatus :one
UPDATE upload_service.media_assets
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    upload_id,
    owner_id,
    asset_type,
    storage_key,
    file_name,
    mime_type,
    size_bytes,
    checksum,
    status,
    created_at,
    updated_at;

-- name: DeleteMediaAsset :exec
DELETE FROM upload_service.media_assets
WHERE id = $1;