CREATE INDEX audit_log_ip_gist
    ON audit_log USING gist (ip inet_ops);