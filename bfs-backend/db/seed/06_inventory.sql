-- Add opening balance for Getränke (50 each)
INSERT INTO inventory_ledger (product_id, delta, reason)
SELECT id, 50, 'opening_balance'
FROM product
WHERE category_id = '01970000-0000-7000-0000-000000000001'
  AND NOT EXISTS (
    SELECT 1 FROM inventory_ledger il
    WHERE il.product_id = product.id AND il.reason = 'opening_balance'
  );

-- Add opening balance for Essen (30 each)
INSERT INTO inventory_ledger (product_id, delta, reason)
SELECT id, 30, 'opening_balance'
FROM product
WHERE category_id = '01970000-0000-7000-0000-000000000002'
  AND NOT EXISTS (
    SELECT 1 FROM inventory_ledger il
    WHERE il.product_id = product.id AND il.reason = 'opening_balance'
  );

-- Add opening balance for Beilage (40 each)
INSERT INTO inventory_ledger (product_id, delta, reason)
SELECT id, 40, 'opening_balance'
FROM product
WHERE category_id = '01970000-0000-7000-0000-000000000003'
  AND NOT EXISTS (
    SELECT 1 FROM inventory_ledger il
    WHERE il.product_id = product.id AND il.reason = 'opening_balance'
  );
