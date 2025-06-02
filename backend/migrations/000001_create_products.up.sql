-- 001_create_products.up.sql
CREATE TABLE products (
  id         SERIAL PRIMARY KEY,
  name       TEXT NOT NULL,
  price      NUMERIC(10,2) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now(),
  deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_products_deleted_at ON products(deleted_at);