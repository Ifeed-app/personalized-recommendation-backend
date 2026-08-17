-- name: CreateUserInterest :one
INSERT INTO user_service.user_interests (
    user_id,
    interest,
    weight
)
VALUES (
           $1,
           $2,
           $3
       )
RETURNING
    user_id,
    interest,
    weight,
    created_at,
    updated_at;

-- name: ListUserInterests :many
SELECT
    user_id,
    interest,
    weight,
    created_at,
    updated_at
FROM user_service.user_interests
WHERE user_id = $1
ORDER BY interest;


-- name: DeleteAllUserInterests :exec
DELETE FROM user_service.user_interests
WHERE user_id = $1;


-- name: UpsertUserInterest :one
INSERT INTO user_service.user_interests (
    user_id,
    interest,
    weight
)
VALUES (
           $1,
           LOWER(TRIM($2)),
           $3
       )
ON CONFLICT (user_id, interest)
    DO UPDATE SET
                  weight = EXCLUDED.weight,
                  updated_at = NOW()
RETURNING
    user_id,
    interest,
    weight,
    created_at,
    updated_at;