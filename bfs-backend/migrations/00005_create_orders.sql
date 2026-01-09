-- +goose Up
-- +goose StatementBegin

-- Orders table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    contact_email VARCHAR(255),
    total_cents BIGINT NOT NULL DEFAULT 0,
    status order_status NOT NULL DEFAULT 'pending',
    origin order_origin DEFAULT 'shop',
    -- Stripe fields
    stripe_session_id VARCHAR(255),
    stripe_payment_intent_id VARCHAR(255),
    stripe_charge_id VARCHAR(255),
    stripe_customer_id VARCHAR(255),
    payment_attempt_id VARCHAR(255),
    -- POS payment details stored as JSONB
    pos_payment JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Order indexes
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_orders_origin ON orders(origin);
CREATE INDEX idx_orders_stripe_payment_intent_id ON orders(stripe_payment_intent_id);

-- Order update trigger
CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Order items table
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    price_per_unit_cents BIGINT NOT NULL DEFAULT 0,
    -- For menu items: reference to parent order item
    parent_item_id UUID REFERENCES order_items(id) ON DELETE CASCADE,
    -- For menu items: which slot this item fills
    menu_slot_id UUID REFERENCES menu_slots(id) ON DELETE SET NULL,
    menu_slot_name VARCHAR(255),
    -- Redemption tracking
    is_redeemed BOOLEAN NOT NULL DEFAULT FALSE,
    redeemed_at TIMESTAMPTZ
);

-- Order item indexes
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
CREATE INDEX idx_order_items_parent_item_id ON order_items(parent_item_id);
CREATE INDEX idx_order_items_is_redeemed ON order_items(is_redeemed);
CREATE INDEX idx_order_items_menu_slot_id ON order_items(menu_slot_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_order_items_menu_slot_id;
DROP INDEX IF EXISTS idx_order_items_is_redeemed;
DROP INDEX IF EXISTS idx_order_items_parent_item_id;
DROP INDEX IF EXISTS idx_order_items_product_id;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP TABLE IF EXISTS order_items;

DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP INDEX IF EXISTS idx_orders_stripe_payment_intent_id;
DROP INDEX IF EXISTS idx_orders_origin;
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_customer_id;
DROP TABLE IF EXISTS orders;

-- +goose StatementEnd
