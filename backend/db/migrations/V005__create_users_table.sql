CREATE TABLE users (
    id nano_id PRIMARY KEY NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash bcrypt_hash NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    is_disabled BOOLEAN DEFAULT FALSE,
    disabled_reason TEXT,
    role_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE RESTRICT
);