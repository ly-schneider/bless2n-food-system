INSERT INTO category (id, name, is_active, position) VALUES
  ('01970000-0000-7000-0000-000000000001', 'Getränke', true, 4),
  ('01970000-0000-7000-0000-000000000002', 'Burgers', true, 2),
  ('01970000-0000-7000-0000-000000000003', 'Beilagen', true, 3),
  ('01970000-0000-7000-0000-000000000004', 'Menüs', true, 1)
ON CONFLICT (id) DO NOTHING;