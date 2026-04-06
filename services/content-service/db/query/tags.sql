-- name: CreateTag :one
INSERT INTO content_service.tags (
    id,
    name
) VALUES (
             $1, $2
         )
RETURNING id, name, created_at;

-- name: GetTagByID :one
SELECT id, name, created_at
FROM content_service.tags
WHERE id = $1;

-- name: GetTagByName :one
SELECT id, name, created_at
FROM content_service.tags
WHERE name = $1;

-- name: ListTags :many
SELECT id, name, created_at
FROM content_service.tags
ORDER BY name ASC;

-- name: AddVideoTag :exec
INSERT INTO content_service.video_tags (
    video_id,
    tag_id
) VALUES (
             $1, $2
         )
ON CONFLICT DO NOTHING;

-- name: ListTagsByVideoID :many
SELECT
    t.id,
    t.name,
    t.created_at
FROM content_service.tags t
         JOIN content_service.video_tags vt
              ON vt.tag_id = t.id
WHERE vt.video_id = $1
ORDER BY t.name ASC;