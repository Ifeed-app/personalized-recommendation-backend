-- name: CreateEmailVerificationToken :one
INSERT INTO user_service.email_verification_tokens (
    id,
    user_id,
    token_hash,
    expires_at
)
VALUES (
           $1,
           $2,
           $3,
           $4
       )
RETURNING
    id,
    user_id,
    token_hash,
    expires_at,
    used_at,
    created_at;


-- name: GetValidEmailVerificationToken :one
SELECT
    id,
    user_id,
    token_hash,
    expires_at,
    used_at,
    created_at
FROM user_service.email_verification_tokens
WHERE token_hash = $1
  AND used_at IS NULL
  AND expires_at > NOW();


-- name: MarkEmailVerificationTokenUsed :one
UPDATE user_service.email_verification_tokens
SET
    used_at = NOW()
WHERE id = $1
  AND used_at IS NULL
RETURNING
    id,
    user_id,
    token_hash,
    expires_at,
    used_at,
    created_at;


-- name: InvalidateUserEmailVerificationTokens :exec
UPDATE user_service.email_verification_tokens
SET
    used_at = NOW()
WHERE user_id = $1
  AND used_at IS NULL;


-- name: DeleteExpiredEmailVerificationTokens :exec
DELETE FROM user_service.email_verification_tokens
WHERE expires_at <= NOW();


-- name: CountActiveEmailVerificationTokens :one
SELECT COUNT(*)
FROM user_service.email_verification_tokens
WHERE user_id = $1
  AND used_at IS NULL
  AND expires_at > NOW();

-- name: GetEmailVerificationTokenByHash :one
SELECT
    id,
    user_id,
    token_hash,
    expires_at,
    used_at,
    created_at
FROM user_service.email_verification_tokens
WHERE token_hash = $1
LIMIT 1;
