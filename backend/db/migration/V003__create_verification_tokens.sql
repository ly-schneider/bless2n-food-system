-- Create verification tokens table
CREATE TABLE verification_tokens (
    id CHAR(14) COLLATE "C" PRIMARY KEY,
    user_id CHAR(14) COLLATE "C",
    token INTEGER NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT chk_verification_tokens_token_length CHECK (LENGTH(token::TEXT) = 6)
);

CREATE INDEX idx_verification_tokens_user_id ON verification_tokens (user_id);
