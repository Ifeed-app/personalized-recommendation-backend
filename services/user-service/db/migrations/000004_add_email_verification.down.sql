DROP INDEX IF EXISTS user_service.idx_email_verification_tokens_expires_at;
DROP INDEX IF EXISTS user_service.idx_email_verification_tokens_user_id;

DROP TABLE IF EXISTS user_service.email_verification_tokens;

ALTER TABLE user_service.users
    DROP COLUMN IF EXISTS email_verified_at,
    DROP COLUMN IF EXISTS email_verified;