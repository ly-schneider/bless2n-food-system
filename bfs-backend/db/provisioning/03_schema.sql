CREATE SCHEMA app AUTHORIZATION app_owner;

GRANT USAGE ON SCHEMA app TO app_runtime, app_readonly;