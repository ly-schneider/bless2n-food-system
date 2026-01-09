-- +goose Up
-- +goose StatementBegin

-- Inventory ledger table (tracks stock changes for products)
CREATE TABLE inventory_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    delta INTEGER NOT NULL,
    reason inventory_reason NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Inventory ledger indexes
CREATE INDEX idx_inventory_ledger_product_id ON inventory_ledger(product_id);
CREATE INDEX idx_inventory_ledger_created_at ON inventory_ledger(created_at);
CREATE INDEX idx_inventory_ledger_product_created ON inventory_ledger(product_id, created_at DESC);

-- Idempotency records table (for idempotent API operations)
CREATE TABLE idempotency_records (
    key VARCHAR(255) PRIMARY KEY,
    scope VARCHAR(255) NOT NULL,
    response JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

-- Idempotency record indexes
CREATE INDEX idx_idempotency_records_scope ON idempotency_records(scope);
CREATE INDEX idx_idempotency_records_expires_at ON idempotency_records(expires_at);
CREATE INDEX idx_idempotency_records_created_at ON idempotency_records(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_idempotency_records_created_at;
DROP INDEX IF EXISTS idx_idempotency_records_expires_at;
DROP INDEX IF EXISTS idx_idempotency_records_scope;
DROP TABLE IF EXISTS idempotency_records;

DROP INDEX IF EXISTS idx_inventory_ledger_product_created;
DROP INDEX IF EXISTS idx_inventory_ledger_created_at;
DROP INDEX IF EXISTS idx_inventory_ledger_product_id;
DROP TABLE IF EXISTS inventory_ledger;

-- +goose StatementEnd
