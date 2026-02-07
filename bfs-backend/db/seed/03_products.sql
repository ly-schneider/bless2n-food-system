INSERT INTO app.product (id, category_id, type, name, image, price_cents, jeton_id, is_active) VALUES
  ('01970000-0000-7000-0002-000000000001', '01970000-0000-7000-0000-000000000001', 'simple', 'Coca Cola', '/assets/images/products/bless2n-takeaway-coca-cola-16x9.png', 350, '01970000-0000-7000-0001-000000000003', true),
  ('01970000-0000-7000-0002-000000000002', '01970000-0000-7000-0000-000000000001', 'simple', 'El Tony Mate', '/assets/images/products/bless2n-takeaway-el-tony-mate-16x9.png', 350, '01970000-0000-7000-0001-000000000002', true),
  ('01970000-0000-7000-0002-000000000003', '01970000-0000-7000-0000-000000000001', 'simple', 'Ice Tea Lemon', '/assets/images/products/bless2n-takeaway-ice-tea-lemon-16x9.png', 350, '01970000-0000-7000-0001-000000000003', true),
  ('01970000-0000-7000-0002-000000000004', '01970000-0000-7000-0000-000000000001', 'simple', 'Wasser', '/assets/images/products/bless2n-takeaway-wasser-prickelnd-16x9.png', 200, '01970000-0000-7000-0001-000000000003', true),
  ('01970000-0000-7000-0002-000000000006', '01970000-0000-7000-0000-000000000001', 'simple', 'Red Bull', '/assets/images/products/bless2n-takeaway-red-bull-16x9.png', 250, '01970000-0000-7000-0001-000000000003', true),
  ('01970000-0000-7000-0002-000000000007', '01970000-0000-7000-0000-000000000001', 'simple', 'Capri Sun', '/assets/images/products/bless2n-takeaway-capri-sun-16x9.png', 150, '01970000-0000-7000-0001-000000000003', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO app.product (id, category_id, type, name, image, price_cents, jeton_id, is_active) VALUES
  ('01970000-0000-7000-0002-000000000010', '01970000-0000-7000-0000-000000000002', 'simple', 'Smash Burger', '/assets/images/products/bless2n-takeaway-smash-burger-16x9.png', 950, '01970000-0000-7000-0001-000000000001', true),
  ('01970000-0000-7000-0002-000000000011', '01970000-0000-7000-0000-000000000002', 'simple', 'Veggie Burger', '/assets/images/products/bless2n-takeaway-veggie-burger-16x9.png', 1050, '01970000-0000-7000-0001-000000000001', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO app.product (id, category_id, type, name, image, price_cents, jeton_id, is_active) VALUES
  ('01970000-0000-7000-0002-000000000020', '01970000-0000-7000-0000-000000000003', 'simple', 'Pommes', '/assets/images/products/bless2n-takeaway-pommes-16x9.png', 450, '01970000-0000-7000-0001-000000000004', true)
ON CONFLICT (id) DO NOTHING;
