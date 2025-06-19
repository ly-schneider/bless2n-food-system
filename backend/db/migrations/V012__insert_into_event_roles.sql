INSERT INTO event_roles (name, display_name, description, is_default, is_active) VALUES
    ('admin', 'Administrator', 'Hat vollen Zugriff auf den Event. Kann Zahlungen ansehen, den Namen ändern, etc.', FALSE, TRUE),
    ('moderator', 'Moderator', 'Hat eingeschränkten Zugriff auf den Event. Kann Mitglieder verwalten, aber z.B. keine Zahlungen ansehen.', FALSE, TRUE),
    ('member', 'Mitglied', 'Hat eingeschränkten Zugriff auf den Event. Kann Analytische Daten einsehen.', TRUE, TRUE);