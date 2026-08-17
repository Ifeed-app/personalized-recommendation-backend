ALTER TABLE user_service.users
    ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN email_verified_at TIMESTAMPTZ;


CREATE TABLE user_service.email_verification_tokens (
                                                        id UUID PRIMARY KEY,
                                                        user_id UUID NOT NULL,
                                                        token_hash TEXT NOT NULL UNIQUE,
                                                        expires_at TIMESTAMPTZ NOT NULL,
                                                        used_at TIMESTAMPTZ,
                                                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                                        CONSTRAINT email_verification_tokens_user_fk
                                                            FOREIGN KEY (user_id)
                                                                REFERENCES user_service.users(id)
                                                                ON DELETE CASCADE
);

CREATE INDEX idx_email_verification_tokens_user_id
    ON user_service.email_verification_tokens(user_id);

CREATE INDEX idx_email_verification_tokens_expires_at
    ON user_service.email_verification_tokens(expires_at);