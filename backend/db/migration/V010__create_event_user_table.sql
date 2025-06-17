CREATE TABLE event_user (    
    event_id nano_id NOT NULL,
    user_id nano_id NOT NULL,
    event_role INTEGER NOT NULL,
    invited_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    joined_at TIMESTAMPTZ,
    FOREIGN KEY (event_id) REFERENCES event (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE,
    FOREIGN KEY (event_role) REFERENCES event_role (id) ON DELETE RESTRICT,
    PRIMARY KEY (event_id, user_id)
);