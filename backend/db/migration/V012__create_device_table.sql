CREATE TABLE device (
    id nano_id PRIMARY KEY NOT NULL,
    serial_number TEXT NOT NULL UNIQUE,
    model TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);