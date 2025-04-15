ALTER TABLE products
  ADD COLUMN category_id UUID REFERENCES categories(id) ON DELETE SET NULL;