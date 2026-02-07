SET search_path TO app, public;

CREATE TABLE "user" (
    id TEXT PRIMARY KEY,
    name TEXT,
    email TEXT UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    image TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Anonymous plugin: marks guest/anonymous users
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    -- Application-specific fields
    role user_role NOT NULL DEFAULT 'customer'
);

CREATE INDEX idx_user_email ON "user" (email);
CREATE INDEX idx_user_is_anonymous ON "user" (is_anonymous) WHERE is_anonymous = TRUE;
CREATE INDEX idx_user_role ON "user" (role);
CREATE INDEX idx_user_created_at ON "user" (created_at DESC);

CREATE TABLE session (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_session_user_id ON session (user_id);
CREATE INDEX idx_session_token ON session (token);
CREATE INDEX idx_session_expires_at ON session (expires_at);

CREATE TABLE account (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    account_id TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    access_token_expires_at TIMESTAMPTZ,
    refresh_token_expires_at TIMESTAMPTZ,
    scope TEXT,
    id_token TEXT,
    password TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider_id, account_id)
);

CREATE INDEX idx_account_user_id ON account (user_id);
CREATE INDEX idx_account_provider_id ON account (provider_id);
CREATE INDEX idx_account_provider_account ON account (provider_id, account_id);

CREATE TABLE verification (
    id TEXT PRIMARY KEY,
    identifier TEXT NOT NULL,
    value TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_verification_identifier ON verification (identifier);
CREATE INDEX idx_verification_expires_at ON verification (expires_at);

CREATE TABLE jwks (
    id TEXT PRIMARY KEY,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jwks_created_at ON jwks (created_at DESC);

CREATE TABLE api_key (
    id TEXT PRIMARY KEY,
    name TEXT,
    start TEXT,
    prefix TEXT,
    key TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    refill_interval INTEGER,
    refill_amount INTEGER,
    last_refill_at TIMESTAMPTZ,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    rate_limit_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    rate_limit_time_window INTEGER,
    rate_limit_max INTEGER,
    request_count INTEGER NOT NULL DEFAULT 0,
    remaining INTEGER,
    last_request TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    permissions TEXT,
    metadata TEXT
);

CREATE INDEX idx_api_key_user_id ON api_key (user_id);
CREATE INDEX idx_api_key_key ON api_key (key);
CREATE INDEX idx_api_key_enabled ON api_key (enabled) WHERE enabled = TRUE;
CREATE INDEX idx_api_key_expires_at ON api_key (expires_at) WHERE expires_at IS NOT NULL;
