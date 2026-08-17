ALTER TABLE user_service.users
    ADD COLUMN birth_year SMALLINT,
    ADD COLUMN gender VARCHAR(30),
    ADD COLUMN country_code CHAR(2),
    ADD COLUMN preferred_language VARCHAR(10),
    ADD COLUMN onboarding_completed BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE user_service.users
    ADD CONSTRAINT users_birth_year_check
        CHECK (
            birth_year IS NULL
                OR birth_year BETWEEN 1900 AND 2100
            ),
    ADD CONSTRAINT users_gender_check
        CHECK (
            gender IS NULL
                OR gender IN (
                              'MALE',
                              'FEMALE',
                              'NON_BINARY',
                              'OTHER'
                )
            ),
    ADD CONSTRAINT users_country_code_check
        CHECK (
            country_code IS NULL
                OR country_code ~ '^[A-Z]{2}$'
            ),
    ADD CONSTRAINT users_preferred_language_check
        CHECK (
            preferred_language IS NULL
                OR preferred_language ~ '^[a-z]{2}(-[A-Z]{2})?$'
            );

CREATE TABLE user_service.user_interests (
                                             user_id UUID NOT NULL,
                                             interest VARCHAR(50) NOT NULL,
                                             weight REAL NOT NULL DEFAULT 1.0,
                                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                             updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                             PRIMARY KEY (user_id, interest),

                                             CONSTRAINT user_interests_user_id_fk
                                                 FOREIGN KEY (user_id)
                                                     REFERENCES user_service.users(id)
                                                     ON DELETE CASCADE,

                                             CONSTRAINT user_interests_interest_check
                                                 CHECK (LENGTH(TRIM(interest)) > 0),

                                             CONSTRAINT user_interests_weight_check
                                                 CHECK (weight >= 0.0 AND weight <= 1.0)
);

CREATE INDEX idx_user_interests_interest
    ON user_service.user_interests(interest);