CREATE INDEX audit_log_created_brin
    ON audit_log USING brin (created_at) WITH (pages_per_range = 32);