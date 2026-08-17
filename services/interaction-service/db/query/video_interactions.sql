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

-- name: GetLatestInteractionByUserAndVideo :one
SELECT
    id,
    user_id,
    video_id,
    interaction_type,
    created_at
FROM integration_service.video_interactions
WHERE user_id = $1
  AND video_id = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: CountCurrentLikesByVideoID :one
SELECT COUNT(*)
FROM (
         SELECT DISTINCT ON (user_id)
             user_id,
             interaction_type
         FROM integration_service.video_interactions
         WHERE video_id = $1
           AND interaction_type IN ('LIKE', 'UNLIKE')
         ORDER BY user_id, created_at DESC
     ) latest_interactions
WHERE interaction_type = 'LIKE';

-- name: IsVideoCurrentlyLikedByUser :one
SELECT COALESCE(
               (
                   SELECT interaction_type = 'LIKE'
                   FROM integration_service.video_interactions
                   WHERE user_id = $1
                     AND video_id = $2
                     AND interaction_type IN ('LIKE', 'UNLIKE')
                   ORDER BY created_at DESC
                   LIMIT 1
               ),
               FALSE
       )::boolean AS liked;

-- name: GetLatestInteractionTypeByUserAndVideo :one
SELECT interaction_type
FROM integration_service.video_interactions
WHERE user_id = $1
  AND video_id = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: GetUserInteractionHistory :many
WITH latest_like_state AS (
    SELECT DISTINCT ON (vi.video_id)
        vi.video_id,
        vi.interaction_type
    FROM integration_service.video_interactions AS vi
    WHERE vi.user_id = $1
      AND vi.interaction_type IN ('LIKE', 'UNLIKE')
    ORDER BY vi.video_id, vi.created_at DESC
),
     skip_summary AS (
         SELECT
             vi.video_id,
             (COUNT(*) > 0)::boolean AS skipped
         FROM integration_service.video_interactions AS vi
         WHERE vi.user_id = $1
           AND vi.interaction_type = 'SKIP'
         GROUP BY vi.video_id
     ),
     watch_summary AS (
         SELECT
             ws.video_id,
             COALESCE(SUM(ws.watched_seconds), 0)::INT AS total_watched_seconds,
             COUNT(*)::INT AS watch_session_count,
             COUNT(*) FILTER (
                 WHERE ws.completed = TRUE
                 )::INT AS completed_count
         FROM integration_service.video_watch_sessions AS ws
         WHERE ws.user_id = $1
         GROUP BY ws.video_id
     ),
     all_videos AS (
         SELECT lls.video_id
         FROM latest_like_state AS lls

         UNION

         SELECT ss.video_id
         FROM skip_summary AS ss

         UNION

         SELECT ws.video_id
         FROM watch_summary AS ws
     )
SELECT
    av.video_id,

    COALESCE(
            lls.interaction_type = 'LIKE',
            FALSE
    )::boolean AS liked,

    COALESCE(
            lls.interaction_type = 'UNLIKE',
            FALSE
    )::boolean AS unliked,

    COALESCE(
            ss.skipped,
            FALSE
    )::boolean AS skipped,

    COALESCE(
            ws.total_watched_seconds,
            0
    )::INT AS total_watched_seconds,

    COALESCE(
            ws.watch_session_count,
            0
    )::INT AS watch_session_count,

    COALESCE(
            ws.completed_count,
            0
    )::INT AS completed_count

FROM all_videos AS av

         LEFT JOIN latest_like_state AS lls
                   ON lls.video_id = av.video_id

         LEFT JOIN skip_summary AS ss
                   ON ss.video_id = av.video_id

         LEFT JOIN watch_summary AS ws
                   ON ws.video_id = av.video_id;