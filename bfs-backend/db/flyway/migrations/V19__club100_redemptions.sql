SET search_path TO app, public;

CREATE TABLE club100_redemption (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    elvanto_person_id VARCHAR(50) NOT NULL,
    elvanto_person_name VARCHAR(100) NOT NULL,
    order_id UUID NOT NULL REFERENCES "order" (id) ON DELETE CASCADE,
    free_product_quantity INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_club100_elvanto_person ON club100_redemption (elvanto_person_id);
