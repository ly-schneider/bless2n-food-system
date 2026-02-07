SET search_path TO app, public;

CREATE TABLE idempotency (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    scope VARCHAR(100) NOT NULL,
    key VARCHAR(100) NOT NULL,
    response JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours',
    UNIQUE (scope, key)
);

CREATE INDEX idx_idempotency_scope_key ON idempotency (scope, key);
CREATE INDEX idx_idempotency_expires_at ON idempotency (expires_at);
