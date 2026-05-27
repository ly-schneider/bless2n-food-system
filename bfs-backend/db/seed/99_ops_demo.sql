-- Realistic ops-dashboard seed.
-- Re-runnable: if any device named 'Grill' already exists, bail.

DO $$
DECLARE
  v_grill_id        UUID := '01970000-0000-7000-1000-000000000001';
  v_pommes_id       UUID := '01970000-0000-7000-1000-000000000002';
  v_drinks_id       UUID := '01970000-0000-7000-1000-000000000003';
  v_dessert_id      UUID := '01970000-0000-7000-1000-000000000004';
  v_pos1_id         UUID := '01970000-0000-7000-1000-000000000010';
  v_pos2_id         UUID := '01970000-0000-7000-1000-000000000011';
  v_pos3_id         UUID := '01970000-0000-7000-1000-000000000012';

  v_p_smash         UUID := '01970000-0000-7000-0002-000000000010';
  v_p_veggie        UUID := '01970000-0000-7000-0002-000000000011';
  v_p_pommes        UUID := '01970000-0000-7000-0002-000000000020';
  v_p_cola          UUID := '01970000-0000-7000-0002-000000000001';
  v_p_eltony        UUID := '01970000-0000-7000-0002-000000000002';
  v_p_icetea        UUID := '01970000-0000-7000-0002-000000000003';
  v_p_wasser        UUID := '01970000-0000-7000-0002-000000000004';
  v_p_redbull       UUID := '01970000-0000-7000-0002-000000000006';
  v_p_capri         UUID := '01970000-0000-7000-0002-000000000007';
  v_p_menu_small    UUID := '01970000-0000-7000-0002-000000000100';
  v_p_menu_large    UUID := '01970000-0000-7000-0002-000000000101';

  v_now             TIMESTAMPTZ := NOW();
  v_order_id        UUID;
  v_line_id         UUID;
  v_created         TIMESTAMPTZ;
  v_delay_min       INT;
  v_qty             INT;
  v_drink           UUID;
  v_burger          UUID;
  i                 INT;
BEGIN
  IF EXISTS (SELECT 1 FROM device WHERE name IN ('Grill','Pommesfritteuse','Getränke','Dessert')) THEN
    RAISE NOTICE 'Demo devices already present — skipping seed.';
    RETURN;
  END IF;

  -- Devices: 3 approved stations + 1 pending station + 2 approved POS + 1 pending POS
  INSERT INTO device (id, name, device_key, type, status, decided_at, created_at) VALUES
    (v_grill_id,   'Grill',            'demo-station-grill',   'STATION', 'approved', v_now - INTERVAL '7 days', v_now - INTERVAL '7 days'),
    (v_pommes_id,  'Pommesfritteuse',  'demo-station-pommes',  'STATION', 'approved', v_now - INTERVAL '7 days', v_now - INTERVAL '7 days'),
    (v_drinks_id,  'Getränke',         'demo-station-drinks',  'STATION', 'approved', v_now - INTERVAL '7 days', v_now - INTERVAL '7 days'),
    (v_dessert_id, 'Dessert',          'demo-station-dessert', 'STATION', 'pending',  NULL,                       v_now - INTERVAL '2 hours'),
    (v_pos1_id,    'POS-Kasse-1',      'demo-pos-1',           'POS',     'approved', v_now - INTERVAL '7 days', v_now - INTERVAL '7 days'),
    (v_pos2_id,    'POS-Kasse-2',      'demo-pos-2',           'POS',     'approved', v_now - INTERVAL '7 days', v_now - INTERVAL '7 days'),
    (v_pos3_id,    'POS-Kasse-3',      'demo-pos-3',           'POS',     'pending',  NULL,                       v_now - INTERVAL '1 hour');

  -- Station product mappings
  INSERT INTO device_product (device_id, product_id) VALUES
    (v_grill_id,  v_p_smash),
    (v_grill_id,  v_p_veggie),
    (v_pommes_id, v_p_pommes),
    (v_drinks_id, v_p_cola),
    (v_drinks_id, v_p_eltony),
    (v_drinks_id, v_p_icetea),
    (v_drinks_id, v_p_wasser),
    (v_drinks_id, v_p_redbull),
    (v_drinks_id, v_p_capri);

  ----------------------------------------------------------------------
  -- GRILL: target YELLOW
  --   1 open simple order (combos contribute ~6 more unredeemed components)
  --  20 redeemed orders across the day, throughput median ~14 min
  ----------------------------------------------------------------------
  FOR i IN 1..1 LOOP
    v_order_id := gen_random_uuid();
    v_created  := v_now - make_interval(mins => 4 + i * 5);  -- 9,14,19,24,29 min ago
    v_burger   := CASE WHEN i % 2 = 0 THEN v_p_veggie ELSE v_p_smash END;

    INSERT INTO "order" (id, total_cents, status, origin, created_at)
      VALUES (v_order_id, 950, 'paid', 'shop', v_created);

    INSERT INTO order_line (id, order_id, line_type, product_id, title, quantity, unit_price_cents)
      VALUES (gen_random_uuid(), v_order_id, 'simple', v_burger,
              CASE WHEN v_burger = v_p_veggie THEN 'Veggie Burger' ELSE 'Smash Burger' END,
              1, 950);
  END LOOP;

  FOR i IN 1..20 LOOP
    v_order_id  := gen_random_uuid();
    v_line_id   := gen_random_uuid();
    v_created   := v_now - make_interval(mins => 35 + i * 12);   -- spread 47m..275m ago
    v_delay_min := 8 + ((i * 7) % 14);                            -- 8..21 min (median ~14)
    v_burger    := CASE WHEN i % 3 = 0 THEN v_p_veggie ELSE v_p_smash END;

    INSERT INTO "order" (id, total_cents, status, origin, created_at)
      VALUES (v_order_id, 950, 'paid', (CASE WHEN i % 4 = 0 THEN 'pos' ELSE 'shop' END)::order_origin, v_created);
    INSERT INTO order_line (id, order_id, line_type, product_id, title, quantity, unit_price_cents)
      VALUES (v_line_id, v_order_id, 'simple', v_burger,
              CASE WHEN v_burger = v_p_veggie THEN 'Veggie Burger' ELSE 'Smash Burger' END,
              1, 950);
    INSERT INTO order_line_redemption (order_line_id, redeemed_at)
      VALUES (v_line_id, v_created + make_interval(mins => v_delay_min));
  END LOOP;

  ----------------------------------------------------------------------
  -- POMMESFRITTEUSE: target RED
  --   10 open orders, ages 8..50 min
  --  15 redeemed, throughput median ~28 min
  ----------------------------------------------------------------------
  FOR i IN 1..10 LOOP
    v_order_id := gen_random_uuid();
    v_created  := v_now - make_interval(mins => 8 + i * 4);
    v_qty      := 1 + (i % 3);

    INSERT INTO "order" (id, total_cents, status, origin, created_at)
      VALUES (v_order_id, 450 * v_qty, 'paid', 'shop', v_created);
    INSERT INTO order_line (id, order_id, line_type, product_id, title, quantity, unit_price_cents)
      VALUES (gen_random_uuid(), v_order_id, 'simple', v_p_pommes, 'Pommes', v_qty, 450);
  END LOOP;

  FOR i IN 1..15 LOOP
    v_order_id  := gen_random_uuid();
    v_line_id   := gen_random_uuid();
    v_created   := v_now - make_interval(mins => 80 + i * 14);
    v_delay_min := 22 + ((i * 5) % 14);

    INSERT INTO "order" (id, total_cents, status, origin, created_at)
      VALUES (v_order_id, 450, 'paid', 'shop', v_created);
    INSERT INTO order_line (id, order_id, line_type, product_id, title, quantity, unit_price_cents)
      VALUES (v_line_id, v_order_id, 'simple', v_p_pommes, 'Pommes', 1, 450);
    INSERT INTO order_line_redemption (order_line_id, redeemed_at)
      VALUES (v_line_id, v_created + make_interval(mins => v_delay_min));
  END LOOP;

  ----------------------------------------------------------------------
  -- GETRÄNKE: target GREEN
  --   1 open order (3 min old)
  --  40 redeemed, throughput median ~3 min
  ----------------------------------------------------------------------
  v_order_id := gen_random_uuid();
  INSERT INTO "order" (id, total_cents, status, origin, created_at)
    VALUES (v_order_id, 350, 'paid', 'shop', v_now - INTERVAL '3 minutes');
  INSERT INTO order_line (id, order_id, line_type, product_id, title, quantity, unit_price_cents)
    VALUES (gen_random_uuid(), v_order_id, 'simple', v_p_cola, 'Coca Cola', 1, 350);

  FOR i IN 1..40 LOOP
    v_order_id  := gen_random_uuid();
    v_line_id   := gen_random_uuid();
    v_created   := v_now - make_interval(mins => 10 + i * 8);
    v_delay_min := 1 + (i % 5);

    -- rotate through drinks
    v_drink := CASE (i % 6)
      WHEN 0 THEN v_p_cola
      WHEN 1 THEN v_p_icetea
      WHEN 2 THEN v_p_eltony
      WHEN 3 THEN v_p_wasser
      WHEN 4 THEN v_p_redbull
      ELSE        v_p_capri
    END;

    INSERT INTO "order" (id, total_cents, status, origin, created_at)
      VALUES (v_order_id, 350, 'paid', 'shop', v_created);
    INSERT INTO order_line (id, order_id, line_type, product_id, title, quantity, unit_price_cents)
      VALUES (v_line_id, v_order_id, 'simple', v_drink,
              CASE v_drink
                WHEN v_p_cola    THEN 'Coca Cola'
                WHEN v_p_icetea  THEN 'Ice Tea Lemon'
                WHEN v_p_eltony  THEN 'El Tony Mate'
                WHEN v_p_wasser  THEN 'Wasser'
                WHEN v_p_redbull THEN 'Red Bull'
                ELSE                  'Capri Sun'
              END,
              1, 350);
    INSERT INTO order_line_redemption (order_line_id, redeemed_at)
      VALUES (v_line_id, v_created + make_interval(mins => v_delay_min));
  END LOOP;

  ----------------------------------------------------------------------
  -- MENU BUNDLES — 6 redeemed combos in the last hour to exercise
  -- bundle/component flattening and "top products last hour"
  ----------------------------------------------------------------------
  FOR i IN 1..6 LOOP
    v_order_id  := gen_random_uuid();
    v_created   := v_now - make_interval(mins => 5 + i * 8);  -- 13..53 min ago

    INSERT INTO "order" (id, total_cents, status, origin, created_at)
      VALUES (v_order_id, 1500, 'paid', 'shop', v_created);

    -- bundle parent (filtered out by flattenOrderLines)
    v_line_id := gen_random_uuid();
    INSERT INTO order_line (id, order_id, line_type, product_id, title, quantity, unit_price_cents)
      VALUES (v_line_id, v_order_id, 'bundle', v_p_menu_large, 'Menü Gross', 1, 1500);

    -- component: Smash Burger -> grill
    INSERT INTO order_line (id, order_id, line_type, product_id, title, quantity, unit_price_cents, parent_line_id)
      VALUES (gen_random_uuid(), v_order_id, 'component', v_p_smash, 'Smash Burger', 1, 0, v_line_id);

    -- component: Pommes -> pommesfritteuse
    INSERT INTO order_line (id, order_id, line_type, product_id, title, quantity, unit_price_cents, parent_line_id)
      VALUES (gen_random_uuid(), v_order_id, 'component', v_p_pommes, 'Pommes', 1, 0, v_line_id);

    -- component: Cola -> drinks (already redeemed quickly)
    v_line_id := gen_random_uuid();
    INSERT INTO order_line (id, order_id, line_type, product_id, title, quantity, unit_price_cents, parent_line_id)
      VALUES (v_line_id, v_order_id, 'component', v_p_cola, 'Coca Cola', 1, 0,
              (SELECT id FROM order_line WHERE order_id = v_order_id AND line_type = 'bundle' LIMIT 1));
    INSERT INTO order_line_redemption (order_line_id, redeemed_at)
      VALUES (v_line_id, v_created + INTERVAL '2 minutes');
  END LOOP;

  ----------------------------------------------------------------------
  -- 5 cancelled orders sprinkled through the day (dashboard's
  -- ops view filters them out via status='paid' — sanity check)
  ----------------------------------------------------------------------
  FOR i IN 1..5 LOOP
    v_order_id := gen_random_uuid();
    v_created  := v_now - make_interval(mins => 40 + i * 30);
    INSERT INTO "order" (id, total_cents, status, origin, created_at)
      VALUES (v_order_id, 1050, 'cancelled', 'shop', v_created);
    INSERT INTO order_line (id, order_id, line_type, product_id, title, quantity, unit_price_cents)
      VALUES (gen_random_uuid(), v_order_id, 'simple', v_p_veggie, 'Veggie Burger', 1, 1050);
  END LOOP;

  RAISE NOTICE 'Ops demo seed complete.';
END $$;
