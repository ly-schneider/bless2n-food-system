SET search_path TO app, public;

ALTER TABLE device_binding ALTER COLUMN id SET DEFAULT uuidv7();
