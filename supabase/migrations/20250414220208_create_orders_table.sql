CREATE TABLE orders (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  admin_id UUID REFERENCES admins(id),  -- Admin who processed/created the order
  order_date TIMESTAMPTZ DEFAULT NOW(),
  status TEXT NOT NULL,  -- e.g., 'pending', 'confirmed', 'shipped'
  total DECIMAL(10,2) DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);