SET search_path TO app, public;

CREATE TABLE device_binding (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_type device_type NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id TEXT NOT NULL REFERENCES "user"(id),
    revoked_at TIMESTAMPTZ,
    station_id UUID REFERENCES device(id)
);

CREATE INDEX idx_device_binding_token_hash ON device_binding (token_hash) WHERE revoked_at IS NULL;
CREATE INDEX idx_device_binding_type ON device_binding (device_type) WHERE revoked_at IS NULL;
CREATE INDEX idx_device_binding_created_by ON device_binding (created_by_user_id);
