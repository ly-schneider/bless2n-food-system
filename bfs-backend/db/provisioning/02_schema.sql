CREATE SCHEMA IF NOT EXISTS app AUTHORIZATION app_owner;
GRANT USAGE ON SCHEMA app TO app_runtime, app_readonly;