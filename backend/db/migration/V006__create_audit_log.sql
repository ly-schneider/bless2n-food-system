-- Create audit log table
CREATE TABLE audit_log (
    id CHAR(14) COLLATE "C" PRIMARY KEY,
    user_id CHAR(14) COLLATE "C",
    ip INET,
    event TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_audit_log_user_id ON audit_log (user_id);
CREATE INDEX idx_audit_log_event ON audit_log (event);
CREATE INDEX idx_audit_log_ip ON audit_log (ip);
