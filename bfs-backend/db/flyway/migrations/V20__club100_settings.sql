SET search_path TO app, public;

ALTER TABLE pos_settings
    ADD COLUMN club100_free_product_id UUID REFERENCES product (id) ON DELETE SET NULL,
    ADD COLUMN club100_max_redemptions INTEGER NOT NULL DEFAULT 2;
