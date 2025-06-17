CREATE TABLE device_pairing (
    event_device_id INTEGER NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('customer', 'cashier')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (event_device_id) REFERENCES event_device (id) ON DELETE CASCADE,
    PRIMARY KEY (event_device_id, role)
);