CREATE TABLE order_items (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
  product_id UUID REFERENCES products(id) ON DELETE CASCADE,
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  price_at_order DECIMAL(10,2) NOT NULL,  -- Records the unit price at the time of ordering
  created_at TIMESTAMPTZ DEFAULT NOW()
);