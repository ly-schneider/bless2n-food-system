CREATE TABLE verification_tokens (
    user_id nano_id NOT NULL,
    token_hash sha256_hash NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, token_hash)
);