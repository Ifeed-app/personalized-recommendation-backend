-- name: GetUserByID :one
SELECT id, email, username, password_hash, created_at, updated_at
FROM user_service.users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, username, password_hash, created_at, updated_at
FROM user_service.users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO user_service.users (
    id,
    email,
    username,
    password_hash
) VALUES (
             $1, $2, $3, $4
         )
RETURNING id, email, username, password_hash, created_at, updated_at;

-- name: UpdateUser :one
UPDATE user_service.users
SET
    email = $2,
    username = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING id, email, username, password_hash, created_at, updated_at;