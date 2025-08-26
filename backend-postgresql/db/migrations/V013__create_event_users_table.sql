CREATE TABLE event_users (    
    event_id nano_id NOT NULL,
    user_id nano_id NOT NULL,
    event_role INTEGER NOT NULL,
    invited_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    joined_at TIMESTAMPTZ,
    FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (event_role) REFERENCES event_roles (id) ON DELETE RESTRICT,
    PRIMARY KEY (event_id, user_id)
);