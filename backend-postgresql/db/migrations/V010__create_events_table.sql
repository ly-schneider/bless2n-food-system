CREATE TABLE events (    
    id nano_id PRIMARY KEY NOT NULL,
    owner_id nano_id NOT NULL,
    name TEXT NOT NULL,
    location TEXT NOT NULL,
    checkout_spots INTEGER NOT NULL CHECK (checkout_spots >= 0),
    is_self_checkout BOOLEAN NOT NULL DEFAULT FALSE,
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (owner_id) REFERENCES users (id) ON DELETE RESTRICT
);