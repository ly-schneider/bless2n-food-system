-- +goose Up
-- +goose StatementBegin

-- Audit logs table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_role VARCHAR(50),
    action audit_action NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    before_state JSONB,
    after_state JSONB,
    request_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit log indexes
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_actor_user_id ON audit_logs(actor_user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);

-- Admin invites table
CREATE TABLE admin_invites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    status admin_invite_status NOT NULL DEFAULT 'pending',
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Admin invite indexes
CREATE UNIQUE INDEX idx_admin_invites_token_hash ON admin_invites(token_hash);
CREATE INDEX idx_admin_invites_invitee_email ON admin_invites(invitee_email);
CREATE INDEX idx_admin_invites_status ON admin_invites(status);
CREATE INDEX idx_admin_invites_expires_at ON admin_invites(expires_at);
CREATE INDEX idx_admin_invites_created_at ON admin_invites(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_admin_invites_created_at;
DROP INDEX IF EXISTS idx_admin_invites_expires_at;
DROP INDEX IF EXISTS idx_admin_invites_status;
DROP INDEX IF EXISTS idx_admin_invites_invitee_email;
DROP INDEX IF EXISTS idx_admin_invites_token_hash;
DROP TABLE IF EXISTS admin_invites;

DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_actor_user_id;
DROP INDEX IF EXISTS idx_audit_logs_entity;
DROP TABLE IF EXISTS audit_logs;

-- +goose StatementEnd
