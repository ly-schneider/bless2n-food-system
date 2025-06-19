CREATE TABLE event_devices (
    id nano_id PRIMARY KEY NOT NULL,
    event_id nano_id NOT NULL,
    device_id nano_id NOT NULL,
    assigned_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE,
    FOREIGN KEY (device_id) REFERENCES devices (id) ON DELETE CASCADE
);