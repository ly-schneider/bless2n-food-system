-- Create verification tokens table
CREATE TABLE verification_tokens (
    user_id UUID REFERENCES users (id) ON DELETE CASCADE,
    token TEXT PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_verification_tokens_user_id ON verification_tokens (user_id);
