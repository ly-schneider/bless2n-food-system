-- +goose Up
-- +goose StatementBegin

-- Categories table
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Category indexes
CREATE UNIQUE INDEX idx_categories_name ON categories(name);
CREATE INDEX idx_categories_is_active ON categories(is_active);
CREATE INDEX idx_categories_position ON categories(position);
CREATE INDEX idx_categories_position_name ON categories(position, name);

-- Category update trigger
CREATE TRIGGER update_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Jetons table (tokens/chips used for menu fulfillment)
CREATE TABLE jetons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    palette_color VARCHAR(50) NOT NULL,
    hex_color VARCHAR(7),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Jeton indexes
CREATE UNIQUE INDEX idx_jetons_name ON jetons(name);

-- Jeton update trigger
CREATE TRIGGER update_jetons_updated_at
    BEFORE UPDATE ON jetons
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_jetons_updated_at ON jetons;
DROP INDEX IF EXISTS idx_jetons_name;
DROP TABLE IF EXISTS jetons;

DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
DROP INDEX IF EXISTS idx_categories_position_name;
DROP INDEX IF EXISTS idx_categories_position;
DROP INDEX IF EXISTS idx_categories_is_active;
DROP INDEX IF EXISTS idx_categories_name;
DROP TABLE IF EXISTS categories;

-- +goose StatementEnd
