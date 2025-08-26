CREATE TABLE event_invites (
    id nano_id PRIMARY KEY NOT NULL,
    event_id nano_id NOT NULL,
    event_role INTEGER NOT NULL,
    invitee_email TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE,
    FOREIGN KEY (event_role) REFERENCES event_roles (id) ON DELETE RESTRICT
);