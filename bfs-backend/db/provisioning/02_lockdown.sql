REVOKE CREATE ON SCHEMA public FROM PUBLIC;
REVOKE CONNECT ON DATABASE bless2n_food_system FROM PUBLIC;
GRANT  CONNECT ON DATABASE bless2n_food_system TO flyway, app_backend, app_admin;