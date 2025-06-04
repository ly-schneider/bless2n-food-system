-- Create password reset tokens table
CREATE TABLE password_reset_tokens (
    user_id UUID REFERENCES users (id) ON DELETE CASCADE,
    token TEXT PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens (user_id);
