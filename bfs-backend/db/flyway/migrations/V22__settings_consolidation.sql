SET search_path TO app, public;

CREATE TABLE settings (
    id VARCHAR(50) PRIMARY KEY DEFAULT 'default',
    pos_mode pos_fulfillment_mode NOT NULL DEFAULT 'QR_CODE',
    club100_max_redemptions INTEGER NOT NULL DEFAULT 2,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE club100_free_product (
    settings_id VARCHAR(50) NOT NULL REFERENCES settings (id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES product (id) ON DELETE CASCADE,
    PRIMARY KEY (settings_id, product_id)
);
