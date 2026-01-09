-- +goose Up
-- +goose StatementBegin

-- Products table
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    type product_type NOT NULL DEFAULT 'simple',
    name VARCHAR(255) NOT NULL,
    image TEXT,
    price_cents BIGINT NOT NULL DEFAULT 0,
    jeton_id UUID REFERENCES jetons(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Product indexes
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_is_active ON products(is_active);
CREATE INDEX idx_products_type ON products(type);
CREATE INDEX idx_products_jeton_id ON products(jeton_id);

-- Product update trigger
CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Menu slots table (slots within a menu product)
CREATE TABLE menu_slots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    sequence INTEGER NOT NULL DEFAULT 0
);

-- Menu slot indexes
CREATE INDEX idx_menu_slots_product_id ON menu_slots(product_id);
CREATE INDEX idx_menu_slots_sequence ON menu_slots(sequence);

-- Menu slot items table (products that can fill a menu slot)
CREATE TABLE menu_slot_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    menu_slot_id UUID NOT NULL REFERENCES menu_slots(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE
);

-- Menu slot item indexes
CREATE INDEX idx_menu_slot_items_menu_slot_id ON menu_slot_items(menu_slot_id);
CREATE INDEX idx_menu_slot_items_product_id ON menu_slot_items(product_id);

-- Unique constraint to prevent duplicate product assignments to a slot
CREATE UNIQUE INDEX idx_menu_slot_items_unique ON menu_slot_items(menu_slot_id, product_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_menu_slot_items_unique;
DROP INDEX IF EXISTS idx_menu_slot_items_product_id;
DROP INDEX IF EXISTS idx_menu_slot_items_menu_slot_id;
DROP TABLE IF EXISTS menu_slot_items;

DROP INDEX IF EXISTS idx_menu_slots_sequence;
DROP INDEX IF EXISTS idx_menu_slots_product_id;
DROP TABLE IF EXISTS menu_slots;

DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP INDEX IF EXISTS idx_products_jeton_id;
DROP INDEX IF EXISTS idx_products_type;
DROP INDEX IF EXISTS idx_products_is_active;
DROP INDEX IF EXISTS idx_products_category_id;
DROP INDEX IF EXISTS idx_products_name;
DROP TABLE IF EXISTS products;

-- +goose StatementEnd
