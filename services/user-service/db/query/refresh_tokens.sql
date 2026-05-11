-- name: CreateRefreshToken :one
INSERT INTO user_service.refresh_tokens (
    id,
    user_id,
    token,
    expires_at
) VALUES (
             $1, $2, $3, $4
         )
RETURNING id, user_id, token, expires_at, created_at, revoked_at;

-- name: GetRefreshTokenByToken :one
SELECT id, user_id, token, expires_at, created_at, revoked_at
FROM user_service.refresh_tokens
WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE user_service.refresh_tokens
SET revoked_at = NOW()
WHERE token = $1;