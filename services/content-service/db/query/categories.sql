-- name: CreateCategory :one
INSERT INTO content_service.categories (
    id,
    name
) VALUES (
             $1, $2
         )
RETURNING id, name, created_at;

-- name: GetCategoryByID :one
SELECT id, name, created_at
FROM content_service.categories
WHERE id = $1;

-- name: GetCategoryByName :one
SELECT id, name, created_at
FROM content_service.categories
WHERE name = $1;

-- name: ListCategories :many
SELECT id, name, created_at
FROM content_service.categories
ORDER BY name ASC;

-- name: AddVideoCategory :exec
INSERT INTO content_service.video_categories (
    video_id,
    category_id
) VALUES (
             $1, $2
         )
ON CONFLICT DO NOTHING;

-- name: ListCategoriesByVideoID :many
SELECT
    c.id,
    c.name,
    c.created_at
FROM content_service.categories c
         JOIN content_service.video_categories vc
              ON vc.category_id = c.id
WHERE vc.video_id = $1
ORDER BY c.name ASC;

-- name: GetOrCreateCategory :one
INSERT INTO content_service.categories (
    id,
    name
) VALUES (
             $1, $2
         )
ON CONFLICT (name)
    DO UPDATE SET name = EXCLUDED.name
RETURNING id, name, created_at;