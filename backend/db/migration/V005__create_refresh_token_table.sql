CREATE TABLE refresh_token (
    user_id nano_id NOT NULL,
    token_hash BYTEA(32) NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, token_hash)
);