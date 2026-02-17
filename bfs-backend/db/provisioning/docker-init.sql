-- Idempotent role creation
DO $$ BEGIN
  CREATE ROLE app_owner NOLOGIN;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE ROLE app_runtime NOLOGIN;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE ROLE app_readonly NOLOGIN;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE ROLE atlas LOGIN NOINHERIT;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE ROLE app_admin LOGIN NOINHERIT;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE ROLE app_backend LOGIN;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Role memberships
GRANT app_runtime  TO app_backend;
GRANT app_owner    TO atlas;
GRANT app_owner    TO app_admin;
GRANT app_owner    TO CURRENT_USER;
GRANT app_runtime  TO CURRENT_USER;
GRANT app_readonly TO CURRENT_USER;

-- Grants on existing objects
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_runtime;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO app_readonly;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO app_runtime;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_readonly;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO app_runtime, app_readonly;

-- Default privileges for future objects
SET ROLE app_owner;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_runtime;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO app_readonly;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO app_runtime;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT USAGE, SELECT ON SEQUENCES TO app_readonly;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT EXECUTE ON FUNCTIONS TO app_runtime, app_readonly;

RESET ROLE;

-- Local dev passwords
ALTER USER atlas       WITH PASSWORD 'atlas';
ALTER USER app_backend WITH PASSWORD 'app_backend';
ALTER USER app_admin   WITH PASSWORD 'app_admin';
