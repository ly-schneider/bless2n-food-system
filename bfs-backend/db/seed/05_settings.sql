INSERT INTO settings (id, pos_mode, club100_max_redemptions) VALUES
  ('default', 'JETON', 2)
ON CONFLICT (id) DO UPDATE SET
  pos_mode = EXCLUDED.pos_mode,
  club100_max_redemptions = EXCLUDED.club100_max_redemptions;

INSERT INTO club100_free_product (settings_id, product_id) VALUES
  ('default', '01970000-0000-7000-0002-000000000010'),  -- Smash Burger
  ('default', '01970000-0000-7000-0002-000000000011')   -- Veggie Burger
ON CONFLICT (settings_id, product_id) DO NOTHING;
