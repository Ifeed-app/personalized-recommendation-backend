-- name: CreateVideoWatchSession :one
INSERT INTO integration_service.video_watch_sessions (
    id,
    user_id,
    video_id,
    watched_seconds,
    completed,
    started_at,
    ended_at
) VALUES (
             $1, $2, $3, $4, $5, $6, $7
         )
RETURNING
    id,
    user_id,
    video_id,
    watched_seconds,
    completed,
    started_at,
    ended_at,
    created_at;

-- name: GetVideoWatchSessionByID :one
SELECT
    id,
    user_id,
    video_id,
    watched_seconds,
    completed,
    started_at,
    ended_at,
    created_at
FROM integration_service.video_watch_sessions
WHERE id = $1;

-- name: ListVideoWatchSessionsByUserID :many
SELECT
    id,
    user_id,
    video_id,
    watched_seconds,
    completed,
    started_at,
    ended_at,
    created_at
FROM integration_service.video_watch_sessions
WHERE user_id = $1
ORDER BY started_at DESC;

-- name: ListVideoWatchSessionsByVideoID :many
SELECT
    id,
    user_id,
    video_id,
    watched_seconds,
    completed,
    started_at,
    ended_at,
    created_at
FROM integration_service.video_watch_sessions
WHERE video_id = $1
ORDER BY started_at DESC;

-- name: ListCompletedWatchSessionsByUserID :many
SELECT
    id,
    user_id,
    video_id,
    watched_seconds,
    completed,
    started_at,
    ended_at,
    created_at
FROM integration_service.video_watch_sessions
WHERE user_id = $1
  AND completed = TRUE
ORDER BY started_at DESC;

-- name: GetTotalWatchedSecondsByUserID :one
SELECT COALESCE(SUM(watched_seconds), 0)
FROM integration_service.video_watch_sessions
WHERE user_id = $1;

-- name: GetTotalWatchedSecondsByVideoID :one
SELECT COALESCE(SUM(watched_seconds), 0)
FROM integration_service.video_watch_sessions
WHERE video_id = $1;

-- name: DeleteVideoWatchSession :exec
DELETE FROM integration_service.video_watch_sessions
WHERE id = $1;