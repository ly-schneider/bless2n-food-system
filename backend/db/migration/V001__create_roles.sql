-- Create roles table
CREATE TABLE roles (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name TEXT UNIQUE NOT NULL
);

CREATE INDEX idx_roles_name ON roles (name);
