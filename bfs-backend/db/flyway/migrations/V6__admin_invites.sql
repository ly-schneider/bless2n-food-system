SET search_path TO app, public;

CREATE TABLE admin_invite (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    invited_by_user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    invitee_email TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'expired', 'revoked')),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_admin_invite_token_hash ON admin_invite (token_hash);
CREATE INDEX idx_admin_invite_status ON admin_invite (status);
CREATE INDEX idx_admin_invite_email ON admin_invite (invitee_email);
