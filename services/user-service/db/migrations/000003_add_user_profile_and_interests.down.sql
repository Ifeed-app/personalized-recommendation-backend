DROP INDEX IF EXISTS user_service.idx_user_interests_interest;

DROP TABLE IF EXISTS user_service.user_interests;

ALTER TABLE user_service.users
    DROP CONSTRAINT IF EXISTS users_preferred_language_check,
    DROP CONSTRAINT IF EXISTS users_country_code_check,
    DROP CONSTRAINT IF EXISTS users_gender_check,
    DROP CONSTRAINT IF EXISTS users_birth_year_check;

ALTER TABLE user_service.users
    DROP COLUMN IF EXISTS onboarding_completed,
    DROP COLUMN IF EXISTS preferred_language,
    DROP COLUMN IF EXISTS country_code,
    DROP COLUMN IF EXISTS gender,
    DROP COLUMN IF EXISTS birth_year;