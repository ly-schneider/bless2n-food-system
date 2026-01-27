ALTER ROLE flyway      SET search_path = app, pg_catalog;
ALTER ROLE app_backend SET search_path = app, pg_catalog;
ALTER ROLE app_admin   SET search_path = app, pg_catalog;