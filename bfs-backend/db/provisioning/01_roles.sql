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

GRANT app_runtime  TO app_backend;
GRANT app_owner    TO atlas;
GRANT app_owner    TO app_admin;
GRANT app_owner    TO CURRENT_USER;
GRANT app_runtime  TO CURRENT_USER;
GRANT app_readonly TO CURRENT_USER;
