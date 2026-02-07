SET search_path TO app, public;

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
    palette_color VARCHAR(20) NOT NULL,
    hex_color VARCHAR(7),
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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_device_device_key ON device (device_key);
CREATE INDEX idx_device_status ON device (status);
CREATE INDEX idx_device_type ON device (type);
CREATE INDEX idx_device_expires_at ON device (expires_at);
CREATE INDEX idx_device_decided_by ON device (decided_by);

CREATE TABLE device_product (
    device_id UUID NOT NULL REFERENCES device (id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES product (id) ON DELETE CASCADE,
    PRIMARY KEY (device_id, product_id)
);

CREATE INDEX idx_device_product_product_id ON device_product (product_id);

CREATE TABLE pos_settings (
    id VARCHAR(50) PRIMARY KEY DEFAULT 'default',
    mode pos_fulfillment_mode NOT NULL DEFAULT 'JETON',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "order" (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    customer_id TEXT REFERENCES "user" (id) ON DELETE SET NULL,
    contact_email VARCHAR(50),
    total_cents BIGINT NOT NULL,
    status order_status NOT NULL,
    origin order_origin NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_customer_id ON "order" (customer_id);
CREATE INDEX idx_order_status ON "order" (status);
CREATE INDEX idx_order_origin ON "order" (origin);
CREATE INDEX idx_order_created_at ON "order" (created_at DESC);
CREATE INDEX idx_order_customer_status ON "order" (customer_id, status);

CREATE TABLE order_payment (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    order_id UUID NOT NULL REFERENCES "order" (id) ON DELETE CASCADE,
    method payment_method NOT NULL,
    amount_cents BIGINT NOT NULL,
    received_cents BIGINT NOT NULL DEFAULT 0,
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
