-- neon_auth schema access (for foreign key references)
GRANT USAGE ON SCHEMA neon_auth TO app_owner, app_runtime;
GRANT REFERENCES ON ALL TABLES IN SCHEMA neon_auth TO app_owner;
GRANT SELECT ON neon_auth."user" TO app_runtime;

-- app schema grants
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA app TO app_runtime;
GRANT SELECT ON ALL TABLES IN SCHEMA app TO app_readonly;

GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA app TO app_runtime;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA app TO app_readonly;

GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA app TO app_runtime, app_readonly;