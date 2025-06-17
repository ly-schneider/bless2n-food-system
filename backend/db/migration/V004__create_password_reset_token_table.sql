CREATE TABLE password_reset_token (
    user_id nano_id NOT NULL,
    token_hash BYTEA(32) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, token_hash)
);