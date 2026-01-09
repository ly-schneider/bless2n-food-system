-- +goose Up
-- +goose StatementBegin

-- Enable UUID extension for generating UUIDs

CREATE TYPE user_role AS ENUM ('admin', 'customer');
CREATE TYPE order_status AS ENUM ('pending', 'paid', 'cancelled', 'refunded');
CREATE TYPE order_origin AS ENUM ('shop', 'pos');
CREATE TYPE product_type AS ENUM ('simple', 'menu');
CREATE TYPE order_item_type AS ENUM ('simple', 'bundle', 'component');
CREATE TYPE inventory_reason AS ENUM ('opening_balance', 'sale', 'refund', 'manual_adjust', 'correction');
CREATE TYPE common_status AS ENUM ('pending', 'approved', 'rejected', 'revoked'); -- POS, Station and Admin Invite common status
CREATE TYPE audit_action AS ENUM ('create', 'update', 'delete');
CREATE TYPE identity_provider AS ENUM ('google');
CREATE TYPE pos_fulfillment_mode AS ENUM ('QR_CODE', 'JETON');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TYPE IF EXISTS pos_fulfillment_mode;
DROP TYPE IF EXISTS identity_provider;
DROP TYPE IF EXISTS audit_action;
DROP TYPE IF EXISTS common_status;
DROP TYPE IF EXISTS inventory_reason;
DROP TYPE IF EXISTS order_item_type;
DROP TYPE IF EXISTS product_type;
DROP TYPE IF EXISTS order_origin;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS user_role;

-- +goose StatementEnd
