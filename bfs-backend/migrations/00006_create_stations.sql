-- +goose Up
-- +goose StatementBegin

-- Stations table (redemption stations/kiosks)
CREATE TABLE stations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    model VARCHAR(255),
    os VARCHAR(255),
    device_key VARCHAR(255) NOT NULL,
    status station_request_status DEFAULT 'pending',
    approved BOOLEAN NOT NULL DEFAULT FALSE,
    approved_at TIMESTAMPTZ,
    decided_by UUID REFERENCES users(id) ON DELETE SET NULL,
    decided_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Station indexes
CREATE UNIQUE INDEX idx_stations_device_key ON stations(device_key);
CREATE INDEX idx_stations_approved ON stations(approved);
CREATE INDEX idx_stations_status ON stations(status);
CREATE INDEX idx_stations_created_at ON stations(created_at);

-- Station update trigger
CREATE TRIGGER update_stations_updated_at
    BEFORE UPDATE ON stations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Station products table (which products a station can redeem)
CREATE TABLE station_products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    station_id UUID NOT NULL REFERENCES stations(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE
);

-- Station product indexes (unique constraint on station_id + product_id)
CREATE UNIQUE INDEX idx_station_products_station_product ON station_products(station_id, product_id);
CREATE INDEX idx_station_products_station_id ON station_products(station_id);
CREATE INDEX idx_station_products_product_id ON station_products(product_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_station_products_product_id;
DROP INDEX IF EXISTS idx_station_products_station_id;
DROP INDEX IF EXISTS idx_station_products_station_product;
DROP TABLE IF EXISTS station_products;

DROP TRIGGER IF EXISTS update_stations_updated_at ON stations;
DROP INDEX IF EXISTS idx_stations_created_at;
DROP INDEX IF EXISTS idx_stations_status;
DROP INDEX IF EXISTS idx_stations_approved;
DROP INDEX IF EXISTS idx_stations_device_key;
DROP TABLE IF EXISTS stations;

-- +goose StatementEnd
