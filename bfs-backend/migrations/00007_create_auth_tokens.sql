-- +goose Up
-- +goose StatementBegin

-- Identity links table (OAuth/OIDC federated identities)
CREATE TABLE identity_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider identity_provider NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    email_snapshot VARCHAR(255),
    display_name VARCHAR(255),
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Identity link indexes
CREATE UNIQUE INDEX idx_identity_links_provider_user ON identity_links(provider, provider_user_id);
CREATE INDEX idx_identity_links_user_id ON identity_links(user_id);
CREATE INDEX idx_identity_links_created_at ON identity_links(created_at);

-- Identity link update trigger
CREATE TRIGGER update_identity_links_updated_at
    BEFORE UPDATE ON identity_links
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Refresh tokens table
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_reason VARCHAR(255),
    family_id VARCHAR(255) NOT NULL
);

-- Refresh token indexes
CREATE UNIQUE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_user_revoked_expires ON refresh_tokens(user_id, is_revoked, expires_at);
CREATE INDEX idx_refresh_tokens_family_revoked ON refresh_tokens(family_id, is_revoked);
CREATE INDEX idx_refresh_tokens_last_used_at ON refresh_tokens(last_used_at);

-- OTP tokens table (one-time passwords for email verification)
CREATE TABLE otp_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ,
    attempts INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL
);

-- OTP token indexes
CREATE INDEX idx_otp_tokens_user_id ON otp_tokens(user_id);
CREATE INDEX idx_otp_tokens_expires_at ON otp_tokens(expires_at);
CREATE INDEX idx_otp_tokens_created_at ON otp_tokens(created_at);

-- Email change tokens table
CREATE TABLE email_change_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    new_email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ,
    attempts INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL
);

-- Email change token indexes
CREATE INDEX idx_email_change_tokens_user_id ON email_change_tokens(user_id);
CREATE INDEX idx_email_change_tokens_new_email ON email_change_tokens(new_email);
CREATE INDEX idx_email_change_tokens_expires_at ON email_change_tokens(expires_at);
CREATE INDEX idx_email_change_tokens_created_at ON email_change_tokens(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_email_change_tokens_created_at;
DROP INDEX IF EXISTS idx_email_change_tokens_expires_at;
DROP INDEX IF EXISTS idx_email_change_tokens_new_email;
DROP INDEX IF EXISTS idx_email_change_tokens_user_id;
DROP TABLE IF EXISTS email_change_tokens;

DROP INDEX IF EXISTS idx_otp_tokens_created_at;
DROP INDEX IF EXISTS idx_otp_tokens_expires_at;
DROP INDEX IF EXISTS idx_otp_tokens_user_id;
DROP TABLE IF EXISTS otp_tokens;

DROP INDEX IF EXISTS idx_refresh_tokens_last_used_at;
DROP INDEX IF EXISTS idx_refresh_tokens_family_revoked;
DROP INDEX IF EXISTS idx_refresh_tokens_user_revoked_expires;
DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;
DROP TABLE IF EXISTS refresh_tokens;

DROP TRIGGER IF EXISTS update_identity_links_updated_at ON identity_links;
DROP INDEX IF EXISTS idx_identity_links_created_at;
DROP INDEX IF EXISTS idx_identity_links_user_id;
DROP INDEX IF EXISTS idx_identity_links_provider_user;
DROP TABLE IF EXISTS identity_links;

-- +goose StatementEnd
