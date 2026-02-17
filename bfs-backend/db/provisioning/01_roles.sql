CREATE ROLE app_owner   NOLOGIN;
CREATE ROLE app_runtime NOLOGIN;
CREATE ROLE app_readonly NOLOGIN;

CREATE ROLE atlas       LOGIN NOINHERIT;
CREATE ROLE app_admin   LOGIN NOINHERIT;

CREATE ROLE app_backend LOGIN;

GRANT app_runtime  TO app_backend;
GRANT app_owner    TO atlas;
GRANT app_owner    TO app_admin;
GRANT app_owner    TO CURRENT_USER;
GRANT app_runtime  TO CURRENT_USER;
GRANT app_readonly TO CURRENT_USER;
