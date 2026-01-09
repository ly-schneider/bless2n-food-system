-- +goose Up
-- +goose StatementBegin

-- POS devices table (tablets/terminals for point of sale)
CREATE TABLE pos_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    model VARCHAR(255),
    os VARCHAR(255),
    device_token VARCHAR(255) NOT NULL,
    status pos_request_status DEFAULT 'pending',
    approved BOOLEAN NOT NULL DEFAULT FALSE,
    approved_at TIMESTAMPTZ,
    decided_by UUID REFERENCES users(id) ON DELETE SET NULL,
    decided_at TIMESTAMPTZ,
    -- Device capabilities
    card_capable BOOLEAN,
    printer_mac VARCHAR(17),
    printer_uuid VARCHAR(36),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- POS device indexes
CREATE UNIQUE INDEX idx_pos_devices_device_token ON pos_devices(device_token);
CREATE INDEX idx_pos_devices_approved ON pos_devices(approved);
CREATE INDEX idx_pos_devices_status ON pos_devices(status);
CREATE INDEX idx_pos_devices_created_at ON pos_devices(created_at);

-- POS device update trigger
CREATE TRIGGER update_pos_devices_updated_at
    BEFORE UPDATE ON pos_devices
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- POS settings table (singleton table for global POS configuration)
CREATE TABLE pos_settings (
    id VARCHAR(50) PRIMARY KEY DEFAULT 'default',
    mode pos_fulfillment_mode NOT NULL DEFAULT 'JETON',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default settings row
INSERT INTO pos_settings (id, mode) VALUES ('default', 'JETON');

-- POS settings update trigger
CREATE TRIGGER update_pos_settings_updated_at
    BEFORE UPDATE ON pos_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_pos_settings_updated_at ON pos_settings;
DROP TABLE IF EXISTS pos_settings;

DROP TRIGGER IF EXISTS update_pos_devices_updated_at ON pos_devices;
DROP INDEX IF EXISTS idx_pos_devices_created_at;
DROP INDEX IF EXISTS idx_pos_devices_status;
DROP INDEX IF EXISTS idx_pos_devices_approved;
DROP INDEX IF EXISTS idx_pos_devices_device_token;
DROP TABLE IF EXISTS pos_devices;

-- +goose StatementEnd
