ALTER TABLE app.device ADD COLUMN pairing_code VARCHAR(6);
ALTER TABLE app.device ADD COLUMN pairing_code_expires_at TIMESTAMPTZ;
CREATE UNIQUE INDEX idx_device_pairing_code ON app.device (pairing_code) WHERE pairing_code IS NOT NULL;
