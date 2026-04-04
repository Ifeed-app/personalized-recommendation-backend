CREATE TABLE IF NOT EXISTS user_service.users (
id UUID PRIMARY KEY,
email TEXT NOT NULL UNIQUE,
username TEXT NOT NULL UNIQUE,
password_hash TEXT NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_service.refresh_tokens (
id UUID PRIMARY KEY,
user_id UUID NOT NULL,
token TEXT NOT NULL UNIQUE,
expires_at TIMESTAMPTZ NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
revoked_at TIMESTAMPTZ,
CONSTRAINT fk_refresh_tokens_user
   FOREIGN KEY (user_id)
       REFERENCES user_service.users(id)
       ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_users_email
    ON user_service.users(email);

CREATE INDEX IF NOT EXISTS idx_users_username
    ON user_service.users(username);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id
    ON user_service.refresh_tokens(user_id);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token
    ON user_service.refresh_tokens(token);