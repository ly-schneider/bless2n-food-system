CREATE TABLE
  app.category (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    name VARCHAR(20) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
  );

CREATE INDEX idx_category_is_active ON app.category (is_active);

CREATE INDEX idx_category_position ON app.category (position);

CREATE TABLE
  app.jeton (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    name VARCHAR(20) NOT NULL,
    palette_color VARCHAR(20) NOT NULL,
    hex_color VARCHAR(7),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
  );

CREATE TABLE
  app.product (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    category_id UUID NOT NULL REFERENCES app.category (id) ON DELETE RESTRICT,
    type product_type NOT NULL DEFAULT 'simple',
    name VARCHAR(20) NOT NULL,
    image TEXT,
    price_cents BIGINT NOT NULL DEFAULT 0,
    jeton_id UUID REFERENCES app.jeton (id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
  );

CREATE INDEX idx_product_category_id ON app.product (category_id);

CREATE INDEX idx_product_jeton_id ON app.product (jeton_id);

CREATE INDEX idx_product_is_active ON app.product (is_active);

CREATE INDEX idx_product_type ON app.product (type);

CREATE INDEX idx_product_category_active ON app.product (category_id, is_active);

CREATE TABLE
  app.menu_slot (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    menu_product_id UUID NOT NULL REFERENCES app.product (id) ON DELETE CASCADE,
    name VARCHAR(20) NOT NULL,
    sequence INTEGER NOT NULL DEFAULT 0
  );

CREATE INDEX idx_menu_slot_menu_product_id ON app.menu_slot (menu_product_id);

CREATE INDEX idx_menu_slot_sequence ON app.menu_slot (sequence);

CREATE TABLE
  app.menu_slot_option (
    menu_slot_id UUID NOT NULL REFERENCES app.menu_slot (id) ON DELETE CASCADE,
    option_product_id UUID NOT NULL REFERENCES app.product (id) ON DELETE RESTRICT,
    PRIMARY KEY (menu_slot_id, option_product_id)
  );

CREATE INDEX idx_menu_slot_option_product_id ON app.menu_slot_option (option_product_id);

CREATE TABLE
  app.device (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    name VARCHAR(20) NOT NULL,
    model VARCHAR(50),
    os VARCHAR(20),
    device_key VARCHAR(100) NOT NULL,
    type device_type NOT NULL,
    status common_status NOT NULL DEFAULT 'pending',
    decided_by UUID REFERENCES neon_auth."user" (id) ON DELETE SET NULL,
    decided_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
  );

CREATE UNIQUE INDEX idx_device_device_key ON app.device (device_key);

CREATE INDEX idx_device_status ON app.device (status);

CREATE INDEX idx_device_type ON app.device (type);

CREATE INDEX idx_device_expires_at ON app.device (expires_at);

CREATE INDEX idx_device_decided_by ON app.device (decided_by);

CREATE TABLE
  app.device_product (
    device_id UUID NOT NULL REFERENCES app.device (id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES app.product (id) ON DELETE CASCADE,
    PRIMARY KEY (device_id, product_id)
  );

CREATE INDEX idx_device_product_product_id ON app.device_product (product_id);

CREATE TABLE
  app.pos_settings (
    id VARCHAR(50) PRIMARY KEY DEFAULT 'default',
    mode pos_fulfillment_mode NOT NULL DEFAULT 'JETON',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
  );

CREATE TABLE
  app.order (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    customer_id UUID REFERENCES neon_auth."user" (id) ON DELETE SET NULL,
    contact_email VARCHAR(50),
    total_cents BIGINT NOT NULL,
    status order_status NOT NULL,
    origin order_origin NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
  );

CREATE INDEX idx_order_customer_id ON app.order (customer_id);

CREATE INDEX idx_order_status ON app.order (status);

CREATE INDEX idx_order_origin ON app.order (origin);

CREATE INDEX idx_order_created_at ON app.order (created_at DESC);

CREATE INDEX idx_order_customer_status ON app.order (customer_id, status);

CREATE TABLE
  app.order_payment (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    order_id UUID NOT NULL REFERENCES app.order (id) ON DELETE CASCADE,
    method payment_method NOT NULL,
    amount_cents BIGINT NOT NULL,
    received_cents BIGINT NOT NULL DEFAULT 0,
    -- POS device used for payment
    device_id UUID REFERENCES app.device (id) ON DELETE SET NULL,
    paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
  );

CREATE INDEX idx_order_payment_order_id ON app.order_payment (order_id);

CREATE INDEX idx_order_payment_device_id ON app.order_payment (device_id);

CREATE INDEX idx_order_payment_method ON app.order_payment (method);

CREATE INDEX idx_order_payment_paid_at ON app.order_payment (paid_at DESC);

CREATE TABLE
  app.order_line (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    order_id UUID NOT NULL REFERENCES app.order (id) ON DELETE CASCADE,
    line_type order_item_type NOT NULL,
    product_id UUID NOT NULL REFERENCES app.product (id) ON DELETE RESTRICT,
    -- Receipt snapshot fields
    title VARCHAR(20) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price_cents BIGINT NOT NULL DEFAULT 0,
    -- Tree structure (bundle/components, menu parent->components)
    parent_line_id UUID REFERENCES app.order_line (id) ON DELETE CASCADE,
    -- Menu linkage + slot-name snapshot
    menu_slot_id UUID REFERENCES app.menu_slot (id) ON DELETE SET NULL,
    menu_slot_name VARCHAR(20)
  );

CREATE INDEX idx_order_line_order_id ON app.order_line (order_id);

CREATE INDEX idx_order_line_product_id ON app.order_line (product_id);

CREATE INDEX idx_order_line_parent_line_id ON app.order_line (parent_line_id);

CREATE INDEX idx_order_line_menu_slot_id ON app.order_line (menu_slot_id);

CREATE TABLE
  app.order_line_redemption (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    order_line_id UUID NOT NULL REFERENCES app.order_line (id) ON DELETE CASCADE,
    redeemed_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
  );

CREATE INDEX idx_order_line_redemption_order_line_id ON app.order_line_redemption (order_line_id);

CREATE INDEX idx_order_line_redemption_redeemed_at ON app.order_line_redemption (redeemed_at DESC);

CREATE TABLE
  app.inventory_ledger (
    id UUID PRIMARY KEY DEFAULT uuidv7 (),
    product_id UUID NOT NULL REFERENCES app.product (id) ON DELETE CASCADE,
    delta INTEGER NOT NULL,
    reason inventory_reason NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    order_id UUID REFERENCES app.order (id) ON DELETE SET NULL,
    order_line_id UUID REFERENCES app.order_line (id) ON DELETE SET NULL,
    device_id UUID REFERENCES app.device (id) ON DELETE SET NULL,
    created_by UUID REFERENCES neon_auth."user" (id) ON DELETE SET NULL
  );

CREATE INDEX idx_inventory_ledger_product_id ON app.inventory_ledger (product_id);

CREATE INDEX idx_inventory_ledger_order_id ON app.inventory_ledger (order_id);

CREATE INDEX idx_inventory_ledger_device_id ON app.inventory_ledger (device_id);

CREATE INDEX idx_inventory_ledger_created_by ON app.inventory_ledger (created_by);

CREATE INDEX idx_inventory_ledger_created_at ON app.inventory_ledger (created_at DESC);

CREATE INDEX idx_inventory_ledger_reason ON app.inventory_ledger (reason);

CREATE INDEX idx_inventory_ledger_product_created ON app.inventory_ledger (product_id, created_at DESC);