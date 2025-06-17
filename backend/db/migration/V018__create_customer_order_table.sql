CREATE TABLE customer_order (
    id nano_id PRIMARY KEY NOT NULL,
    event_id nano_id NOT NULL,
    device_id nano_id NOT NULL,
    total NUMERIC(6, 2) NOT NULL CHECK (total >= 0),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'cancelled')),
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    FOREIGN KEY (event_id) REFERENCES event (id) ON DELETE CASCADE,
    FOREIGN KEY (device_id) REFERENCES event_device (id) ON DELETE CASCADE
);