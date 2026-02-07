ALTER TABLE app.jeton ADD COLUMN color VARCHAR(7) NOT NULL;
ALTER TABLE app.jeton DROP COLUMN palette_color;
ALTER TABLE app.jeton DROP COLUMN hex_color;
