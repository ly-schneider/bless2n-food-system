CREATE TABLE event_product (
    id nano_id PRIMARY KEY NOT NULL,
    event_id nano_id NOT NULL,
    event_category_id nano_id NOT NULL,
    name VARCHAR(30) NOT NULL,
    emoji TEXT CHECK (char_length(emoji) = 1) CHECK (octet_length(emoji) <= 35), -- Ensure input is a single emoji character,
    price NUMERIC(6, 2) NOT NULL CHECK (price >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (event_id) REFERENCES event (id) ON DELETE CASCADE,
    FOREIGN KEY (event_category_id) REFERENCES event_category (id) ON DELETE CASCADE
);