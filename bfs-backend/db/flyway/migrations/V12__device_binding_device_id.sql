ALTER TABLE app.device_binding ADD COLUMN device_id UUID REFERENCES app.device(id);
CREATE INDEX idx_device_binding_device_id ON app.device_binding (device_id) WHERE revoked_at IS NULL;
