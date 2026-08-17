-- name: GetUserByID :one
SELECT
    id,
    email,
    username,
    password_hash,
    birth_year,
    gender,
    country_code,
    preferred_language,
    onboarding_completed,
    email_verified,
    email_verified_at,
    created_at,
    updated_at
FROM user_service.users
WHERE id = $1;


-- name: GetUserByEmail :one
SELECT
    id,
    email,
    username,
    password_hash,
    birth_year,
    gender,
    country_code,
    preferred_language,
    onboarding_completed,
    email_verified,
    email_verified_at,
    created_at,
    updated_at
FROM user_service.users
WHERE LOWER(email) = LOWER(TRIM($1));


-- name: GetUserByUsername :one
SELECT
    id,
    email,
    username,
    password_hash,
    birth_year,
    gender,
    country_code,
    preferred_language,
    onboarding_completed,
    email_verified,
    email_verified_at,
    created_at,
    updated_at
FROM user_service.users
WHERE LOWER(username) = LOWER(TRIM($1));


-- name: GetUserByLoginIdentifier :one
SELECT
    id,
    email,
    username,
    password_hash,
    birth_year,
    gender,
    country_code,
    preferred_language,
    onboarding_completed,
    email_verified,
    email_verified_at,
    created_at,
    updated_at
FROM user_service.users
WHERE LOWER(email) = LOWER($1)
   OR LOWER(username) = LOWER($1)
LIMIT 1;

-- name: CreateUser :one
INSERT INTO user_service.users (
    id,
    email,
    username,
    password_hash,
    birth_year,
    gender,
    country_code,
    preferred_language,
    onboarding_completed,
    email_verified
)
VALUES (
           $1,
           $2,
           $3,
           $4,
           $5,
           $6,
           $7,
           $8,
           FALSE,
           FALSE
       )
RETURNING
    id,
    email,
    username,
    password_hash,
    birth_year,
    gender,
    country_code,
    preferred_language,
    onboarding_completed,
    email_verified,
    email_verified_at,
    created_at,
    updated_at;


-- name: CompleteUserOnboarding :one
UPDATE user_service.users
SET
    onboarding_completed = TRUE,
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    email,
    username,
    password_hash,
    birth_year,
    gender,
    country_code,
    preferred_language,
    onboarding_completed,
    email_verified,
    email_verified_at,
    created_at,
    updated_at;


-- name: UpdateUserProfile :one
UPDATE user_service.users
SET
    username = TRIM($2),
    birth_year = $3,
    gender = $4,
    country_code = UPPER(TRIM($5)),
    preferred_language = LOWER(TRIM($6)),
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    email,
    username,
    password_hash,
    birth_year,
    gender,
    country_code,
    preferred_language,
    onboarding_completed,
    email_verified,
    email_verified_at,
    created_at,
    updated_at;


-- name: MarkEmailVerified :one
UPDATE user_service.users
SET
    email_verified = TRUE,
    email_verified_at = COALESCE(email_verified_at, NOW()),
    updated_at = NOW()
WHERE id = $1
RETURNING
    id,
    email,
    username,
    password_hash,
    birth_year,
    gender,
    country_code,
    preferred_language,
    onboarding_completed,
    email_verified,
    email_verified_at,
    created_at,
    updated_at;


-- name: IsEmailVerified :one
SELECT email_verified
FROM user_service.users
WHERE id = $1;

-- name: MarkUserEmailVerified :exec
UPDATE user_service.users
SET
    email_verified = TRUE,
    email_verified_at = NOW(),
    updated_at = NOW()
WHERE id = $1;