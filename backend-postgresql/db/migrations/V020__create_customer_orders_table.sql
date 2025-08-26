CREATE TABLE customer_orders (
    id nano_id PRIMARY KEY NOT NULL,
    event_id nano_id NOT NULL,
    event_device_id nano_id NOT NULL,
    total NUMERIC(6, 2) NOT NULL CHECK (total >= 0),
    status order_status_enum NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE,
    FOREIGN KEY (event_device_id) REFERENCES event_devices (id) ON DELETE CASCADE
);