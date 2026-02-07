-- Add opening balance for Getränke (50 each)
INSERT INTO app.inventory_ledger (product_id, delta, reason)
SELECT id, 50, 'opening_balance'
FROM app.product
WHERE category_id = '01970000-0000-7000-0000-000000000001'
  AND NOT EXISTS (
    SELECT 1 FROM app.inventory_ledger il
    WHERE il.product_id = app.product.id AND il.reason = 'opening_balance'
  );

-- Add opening balance for Essen (30 each)
INSERT INTO app.inventory_ledger (product_id, delta, reason)
SELECT id, 30, 'opening_balance'
FROM app.product
WHERE category_id = '01970000-0000-7000-0000-000000000002'
  AND NOT EXISTS (
    SELECT 1 FROM app.inventory_ledger il
    WHERE il.product_id = app.product.id AND il.reason = 'opening_balance'
  );

-- Add opening balance for Beilage (40 each)
INSERT INTO app.inventory_ledger (product_id, delta, reason)
SELECT id, 40, 'opening_balance'
FROM app.product
WHERE category_id = '01970000-0000-7000-0000-000000000003'
  AND NOT EXISTS (
    SELECT 1 FROM app.inventory_ledger il
    WHERE il.product_id = app.product.id AND il.reason = 'opening_balance'
  );
