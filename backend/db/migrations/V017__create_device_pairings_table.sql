CREATE TABLE device_pairings (
    id nano_id PRIMARY KEY NOT NULL,
    event_device_id nano_id NOT NULL,
    role device_pairing_role_enum NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (event_device_id) REFERENCES event_devices (id) ON DELETE CASCADE
);