-- +goose Up
-- +goose StatementBegin

-- Enable UUID extension for generating UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- User roles enum
CREATE TYPE user_role AS ENUM ('admin', 'customer');

-- Order status enum
CREATE TYPE order_status AS ENUM ('pending', 'paid', 'cancelled', 'refunded');

-- Order origin enum
CREATE TYPE order_origin AS ENUM ('shop', 'pos');

-- Product type enum
CREATE TYPE product_type AS ENUM ('simple', 'menu');

-- Order item type enum
CREATE TYPE order_item_type AS ENUM ('simple', 'bundle', 'component');

-- Inventory reason enum
CREATE TYPE inventory_reason AS ENUM ('opening_balance', 'sale', 'refund', 'manual_adjust', 'correction');

-- Station request status enum
CREATE TYPE station_request_status AS ENUM ('pending', 'approved', 'rejected', 'revoked');

-- POS request status enum (same values as station but separate type for clarity)
CREATE TYPE pos_request_status AS ENUM ('pending', 'approved', 'rejected', 'revoked');

-- Admin invite status enum
CREATE TYPE admin_invite_status AS ENUM ('pending', 'accepted', 'expired', 'revoked');

-- Audit action enum
CREATE TYPE audit_action AS ENUM ('create', 'update', 'delete');

-- Identity provider enum
CREATE TYPE identity_provider AS ENUM ('google');

-- POS fulfillment mode enum
CREATE TYPE pos_fulfillment_mode AS ENUM ('QR_CODE', 'JETON');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TYPE IF EXISTS pos_fulfillment_mode;
DROP TYPE IF EXISTS identity_provider;
DROP TYPE IF EXISTS audit_action;
DROP TYPE IF EXISTS admin_invite_status;
DROP TYPE IF EXISTS pos_request_status;
DROP TYPE IF EXISTS station_request_status;
DROP TYPE IF EXISTS inventory_reason;
DROP TYPE IF EXISTS order_item_type;
DROP TYPE IF EXISTS product_type;
DROP TYPE IF EXISTS order_origin;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS user_role;

DROP EXTENSION IF EXISTS "uuid-ossp";

-- +goose StatementEnd
