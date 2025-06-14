-- Create password reset tokens table
CREATE TABLE password_reset_tokens (
    id CHAR(14) COLLATE "C" PRIMARY KEY,
    user_id CHAR(14) COLLATE "C",
    token UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    expires_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens (user_id);
