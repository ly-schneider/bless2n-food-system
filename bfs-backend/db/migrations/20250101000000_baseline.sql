-- Baseline migration: consolidated from Flyway V1–V24.
-- For existing databases, apply with: atlas migrate apply --baseline 20250101000000

CREATE SCHEMA IF NOT EXISTS app;
SET search_path TO app, public;

-- ============================================================
-- Enum types
-- ============================================================

CREATE TYPE user_role AS ENUM ('admin', 'customer');

CREATE TYPE order_status AS ENUM ('pending', 'paid', 'cancelled', 'refunded');

CREATE TYPE order_origin AS ENUM ('shop', 'pos');

CREATE TYPE product_type AS ENUM ('simple', 'menu');

CREATE TYPE order_item_type AS ENUM ('simple', 'bundle', 'component');

CREATE TYPE inventory_reason AS ENUM (
  'opening_balance',
  'sale',
  'refund',
  'cancellation',
  'manual_adjust',
  'correction'
);

CREATE TYPE common_status AS ENUM ('pending', 'approved', 'rejected', 'revoked');

CREATE TYPE pos_fulfillment_mode AS ENUM ('QR_CODE', 'JETON');

CREATE TYPE payment_method AS ENUM ('CASH', 'CARD', 'TWINT', 'GRATIS_GUEST', 'GRATIS_VIP', 'GRATIS_100CLUB', 'GRATIS_STAFF');

CREATE TYPE device_type AS ENUM ('POS', 'STATION');

-- ============================================================
-- Better Auth tables
-- ============================================================

CREATE TABLE "user" (
    id TEXT PRIMARY KEY,
    name TEXT,
    email TEXT UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    image TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    role user_role NOT NULL DEFAULT 'customer',
    banned BOOLEAN NOT NULL DEFAULT FALSE,
    ban_reason TEXT,
    ban_expires TIMESTAMPTZ,
    is_club_100 BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_user_email ON "user" (email);
CREATE INDEX idx_user_is_anonymous ON "user" (is_anonymous) WHERE is_anonymous = TRUE;
CREATE INDEX idx_user_role ON "user" (role);
CREATE INDEX idx_user_created_at ON "user" (created_at DESC);

CREATE TABLE session (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    impersonated_by TEXT
);

CREATE INDEX idx_session_user_id ON session (user_id);
CREATE INDEX idx_session_token ON session (token);
CREATE INDEX idx_session_expires_at ON session (expires_at);

CREATE TABLE account (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    account_id TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    access_token_expires_at TIMESTAMPTZ,
    refresh_token_expires_at TIMESTAMPTZ,
    scope TEXT,
    id_token TEXT,
    password TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider_id, account_id)
);

CREATE INDEX idx_account_user_id ON account (user_id);
CREATE INDEX idx_account_provider_id ON account (provider_id);
CREATE INDEX idx_account_provider_account ON account (provider_id, account_id);

CREATE TABLE verification (
    id TEXT PRIMARY KEY,
    identifier TEXT NOT NULL,
    value TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_verification_identifier ON verification (identifier);
CREATE INDEX idx_verification_expires_at ON verification (expires_at);

CREATE TABLE jwks (
    id TEXT PRIMARY KEY,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_jwks_created_at ON jwks (created_at DESC);

CREATE TABLE api_key (
    id TEXT PRIMARY KEY,
    name TEXT,
    start TEXT,
    prefix TEXT,
    key TEXT NOT NULL,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    refill_interval INTEGER,
    refill_amount INTEGER,
    last_refill_at TIMESTAMPTZ,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    rate_limit_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    rate_limit_time_window INTEGER,
    rate_limit_max INTEGER,
    request_count INTEGER NOT NULL DEFAULT 0,
    remaining INTEGER,
    last_request TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    permissions TEXT,
    metadata TEXT
);

CREATE INDEX idx_api_key_user_id ON api_key (user_id);
CREATE INDEX idx_api_key_key ON api_key (key);
CREATE INDEX idx_api_key_enabled ON api_key (enabled) WHERE enabled = TRUE;
CREATE INDEX idx_api_key_expires_at ON api_key (expires_at) WHERE expires_at IS NOT NULL;

-- ============================================================
-- Application tables
-- ============================================================

CREATE TABLE category (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(20) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_category_is_active ON category (is_active);
CREATE INDEX idx_category_position ON category (position);

CREATE TABLE jeton (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(20) NOT NULL,
    color VARCHAR(7) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE product (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    category_id UUID NOT NULL REFERENCES category (id) ON DELETE RESTRICT,
    type product_type NOT NULL DEFAULT 'simple',
    name VARCHAR(20) NOT NULL,
    image TEXT,
    price_cents BIGINT NOT NULL DEFAULT 0,
    jeton_id UUID REFERENCES jeton (id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_category_id ON product (category_id);
CREATE INDEX idx_product_jeton_id ON product (jeton_id);
CREATE INDEX idx_product_is_active ON product (is_active);
CREATE INDEX idx_product_type ON product (type);
CREATE INDEX idx_product_category_active ON product (category_id, is_active);

CREATE TABLE menu_slot (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    menu_product_id UUID NOT NULL REFERENCES product (id) ON DELETE CASCADE,
    name VARCHAR(20) NOT NULL,
    sequence INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_menu_slot_menu_product_id ON menu_slot (menu_product_id);
CREATE INDEX idx_menu_slot_sequence ON menu_slot (sequence);

CREATE TABLE menu_slot_option (
    menu_slot_id UUID NOT NULL REFERENCES menu_slot (id) ON DELETE CASCADE,
    option_product_id UUID NOT NULL REFERENCES product (id) ON DELETE RESTRICT,
    PRIMARY KEY (menu_slot_id, option_product_id)
);

CREATE INDEX idx_menu_slot_option_product_id ON menu_slot_option (option_product_id);

CREATE TABLE device (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(20) NOT NULL,
    model VARCHAR(50),
    os VARCHAR(20),
    device_key VARCHAR(100) NOT NULL,
    type device_type NOT NULL,
    status common_status NOT NULL DEFAULT 'pending',
    decided_by TEXT REFERENCES "user" (id) ON DELETE SET NULL,
    decided_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    pending_session_token TEXT,
    pairing_code VARCHAR(6),
    pairing_code_expires_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_device_device_key ON device (device_key);
CREATE INDEX idx_device_status ON device (status);
CREATE INDEX idx_device_type ON device (type);
CREATE INDEX idx_device_expires_at ON device (expires_at);
CREATE INDEX idx_device_decided_by ON device (decided_by);
CREATE UNIQUE INDEX idx_device_pairing_code ON device (pairing_code) WHERE pairing_code IS NOT NULL;

CREATE TABLE device_product (
    device_id UUID NOT NULL REFERENCES device (id) ON DELETE RESTRICT,
    product_id UUID NOT NULL REFERENCES product (id) ON DELETE CASCADE,
    PRIMARY KEY (device_id, product_id)
);

CREATE INDEX idx_device_product_product_id ON device_product (product_id);

CREATE TABLE device_binding (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    device_type device_type NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id TEXT NOT NULL REFERENCES "user"(id),
    revoked_at TIMESTAMPTZ,
    station_id UUID REFERENCES device(id) ON DELETE CASCADE,
    device_id UUID REFERENCES device(id) ON DELETE CASCADE
);

CREATE INDEX idx_device_binding_token_hash ON device_binding (token_hash) WHERE revoked_at IS NULL;
CREATE INDEX idx_device_binding_type ON device_binding (device_type) WHERE revoked_at IS NULL;
CREATE INDEX idx_device_binding_created_by ON device_binding (created_by_user_id);
CREATE INDEX idx_device_binding_device_id ON device_binding (device_id) WHERE revoked_at IS NULL;

CREATE TABLE "order" (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    customer_id TEXT REFERENCES "user" (id) ON DELETE SET NULL,
    contact_email VARCHAR(50),
    total_cents BIGINT NOT NULL,
    status order_status NOT NULL,
    origin order_origin NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payment_attempt_id VARCHAR(100),
    payrexx_gateway_id INTEGER,
    payrexx_transaction_id INTEGER
);

CREATE INDEX idx_order_customer_id ON "order" (customer_id);
CREATE INDEX idx_order_status ON "order" (status);
CREATE INDEX idx_order_origin ON "order" (origin);
CREATE INDEX idx_order_created_at ON "order" (created_at DESC);
CREATE INDEX idx_order_customer_status ON "order" (customer_id, status);
CREATE UNIQUE INDEX idx_order_payment_attempt_pending
  ON "order" (payment_attempt_id)
  WHERE status = 'pending' AND payment_attempt_id IS NOT NULL;
CREATE INDEX idx_order_payrexx_gateway ON "order" (payrexx_gateway_id) WHERE payrexx_gateway_id IS NOT NULL;
CREATE INDEX idx_order_payrexx_transaction ON "order" (payrexx_transaction_id) WHERE payrexx_transaction_id IS NOT NULL;

CREATE TABLE order_payment (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    order_id UUID NOT NULL REFERENCES "order" (id) ON DELETE CASCADE,
    method payment_method NOT NULL,
    amount_cents BIGINT NOT NULL,
    device_id UUID REFERENCES device (id) ON DELETE SET NULL,
    paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_payment_order_id ON order_payment (order_id);
CREATE INDEX idx_order_payment_device_id ON order_payment (device_id);
CREATE INDEX idx_order_payment_method ON order_payment (method);
CREATE INDEX idx_order_payment_paid_at ON order_payment (paid_at DESC);

CREATE TABLE order_line (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    order_id UUID NOT NULL REFERENCES "order" (id) ON DELETE CASCADE,
    line_type order_item_type NOT NULL,
    product_id UUID NOT NULL REFERENCES product (id) ON DELETE RESTRICT,
    title VARCHAR(20) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price_cents BIGINT NOT NULL DEFAULT 0,
    parent_line_id UUID REFERENCES order_line (id) ON DELETE CASCADE,
    menu_slot_id UUID REFERENCES menu_slot (id) ON DELETE SET NULL,
    menu_slot_name VARCHAR(20)
);

CREATE INDEX idx_order_line_order_id ON order_line (order_id);
CREATE INDEX idx_order_line_product_id ON order_line (product_id);
CREATE INDEX idx_order_line_parent_line_id ON order_line (parent_line_id);
CREATE INDEX idx_order_line_menu_slot_id ON order_line (menu_slot_id);

CREATE TABLE order_line_redemption (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    order_line_id UUID NOT NULL REFERENCES order_line (id) ON DELETE CASCADE,
    redeemed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_line_redemption_order_line_id ON order_line_redemption (order_line_id);
CREATE INDEX idx_order_line_redemption_redeemed_at ON order_line_redemption (redeemed_at DESC);

CREATE TABLE inventory_ledger (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    product_id UUID NOT NULL REFERENCES product (id) ON DELETE CASCADE,
    delta INTEGER NOT NULL,
    reason inventory_reason NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    order_id UUID REFERENCES "order" (id) ON DELETE SET NULL,
    order_line_id UUID REFERENCES order_line (id) ON DELETE SET NULL,
    device_id UUID REFERENCES device (id) ON DELETE SET NULL,
    created_by TEXT REFERENCES "user" (id) ON DELETE SET NULL
);

CREATE INDEX idx_inventory_ledger_product_id ON inventory_ledger (product_id);
CREATE INDEX idx_inventory_ledger_order_id ON inventory_ledger (order_id);
CREATE INDEX idx_inventory_ledger_device_id ON inventory_ledger (device_id);
CREATE INDEX idx_inventory_ledger_created_by ON inventory_ledger (created_by);
CREATE INDEX idx_inventory_ledger_created_at ON inventory_ledger (created_at DESC);
CREATE INDEX idx_inventory_ledger_reason ON inventory_ledger (reason);
CREATE INDEX idx_inventory_ledger_product_created ON inventory_ledger (product_id, created_at DESC);

CREATE TABLE idempotency (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    scope VARCHAR(100) NOT NULL,
    key VARCHAR(100) NOT NULL,
    response JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours',
    UNIQUE (scope, key)
);

CREATE INDEX idx_idempotency_scope_key ON idempotency (scope, key);
CREATE INDEX idx_idempotency_expires_at ON idempotency (expires_at);

CREATE TABLE admin_invite (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    invited_by_user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    invitee_email TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'expired', 'revoked')),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_admin_invite_token_hash ON admin_invite (token_hash);
CREATE INDEX idx_admin_invite_status ON admin_invite (status);
CREATE INDEX idx_admin_invite_email ON admin_invite (invitee_email);

CREATE TABLE club100_redemption (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    elvanto_person_id VARCHAR(50) NOT NULL,
    elvanto_person_name VARCHAR(100) NOT NULL,
    order_id UUID NOT NULL REFERENCES "order" (id) ON DELETE CASCADE,
    free_product_quantity INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_club100_elvanto_person ON club100_redemption (elvanto_person_id);

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
