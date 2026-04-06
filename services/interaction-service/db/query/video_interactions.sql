-- name: CreateVideoInteraction :one
INSERT INTO integration_service.video_interactions (
    id,
    user_id,
    video_id,
    interaction_type
) VALUES (
             $1, $2, $3, $4
         )
RETURNING id, user_id, video_id, interaction_type, created_at;

-- name: ListVideoInteractionsByUserID :many
SELECT
    id,
    user_id,
    video_id,
    interaction_type,
    created_at
FROM integration_service.video_interactions
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListVideoInteractionsByVideoID :many
SELECT
    id,
    user_id,
    video_id,
    interaction_type,
    created_at
FROM integration_service.video_interactions
WHERE video_id = $1
ORDER BY created_at DESC;

-- name: ListVideoInteractionsByUserIDAndType :many
SELECT
    id,
    user_id,
    video_id,
    interaction_type,
    created_at
FROM integration_service.video_interactions
WHERE user_id = $1
  AND interaction_type = $2
ORDER BY created_at DESC;

-- name: CountVideoInteractionsByVideoIDAndType :one
SELECT COUNT(*)
FROM integration_service.video_interactions
WHERE video_id = $1
  AND interaction_type = $2;

-- name: DeleteVideoInteraction :exec
DELETE FROM integration_service.video_interactions
WHERE id = $1;