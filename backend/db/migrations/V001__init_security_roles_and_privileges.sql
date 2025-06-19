/* ----------  V1 : roles, privileges, default grants  ---------- */

-- 1. two login roles -------------------------------------------------
CREATE ROLE ${DB_ADMIN_USER} LOGIN
  PASSWORD '${DB_ADMIN_PASSWORD}'
  BYPASSRLS;                     -- new in PG-17  🡲 super-easy admin 🔒
-- ref: PostgreSQL 17 docs on BYPASSRLS:contentReference[oaicite:3]{index=3}

CREATE ROLE ${APP_USER}  LOGIN
  PASSWORD '${APP_USER_PASSWORD}';

-- 2. make db_admin own the database ---------------------------------
ALTER DATABASE rentro OWNER TO ${DB_ADMIN_USER};

-- 3. baseline privileges (existing + future objects) ----------------
GRANT CONNECT ON DATABASE rentro            TO ${APP_USER};
GRANT USAGE   ON SCHEMA  public             TO ${APP_USER};
GRANT SELECT,INSERT,UPDATE,DELETE
      ON ALL TABLES IN SCHEMA public        TO ${APP_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA public
      GRANT SELECT,INSERT,UPDATE,DELETE ON TABLES TO ${APP_USER};

-- helper GUC so the application can stamp the logical user id
CREATE FUNCTION set_app_user_id(p text) RETURNS void
  LANGUAGE plpgsql SECURITY DEFINER AS
$$BEGIN
  PERFORM set_config('app.current_user_id', p, true);
END$$;

GRANT EXECUTE ON FUNCTION set_app_user_id(text) TO ${APP_USER};

-- 4. make db_admin own the public schema ---------------------------
ALTER SCHEMA public OWNER TO db_admin;
GRANT USAGE ON SCHEMA public TO db_admin;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO db_admin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
       GRANT ALL ON TABLES TO db_admin;