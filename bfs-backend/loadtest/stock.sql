-- Bump inventory for every active simple product so a full baseline loadtest
-- run can't deplete stock. Inventory is a ledger (sum of deltas) — each run
-- of this script adds +50,000 units per product, tagged with created_by =
-- 'loadtest_user_admin_001' so cleanup.sql can remove just the loadtest-
-- contributed stock later without touching real opening balances.
--
-- USAGE
--   psql "$DATABASE_URL" -f loadtest/stock.sql
--
-- Sizing rationale
--   baseline = 7 RPS customer + 3 RPS station × ~11 min × 1-3 items each ≈
--   14k item-units consumed. 50k per product across ~12 products = 600k
--   units of headroom, survives 40+ baseline runs before depletion.

\set ON_ERROR_STOP on

BEGIN;

INSERT INTO inventory_ledger (product_id, delta, reason, created_by)
SELECT id, 50000, 'manual_adjust', 'loadtest_user_admin_001'
FROM product
WHERE type = 'simple' AND is_active = TRUE;

COMMIT;

\echo ''
\echo 'Stock summary after bump:'
SELECT
  p.name,
  COALESCE(SUM(il.delta), 0) AS stock
FROM product p
LEFT JOIN inventory_ledger il ON il.product_id = p.id
WHERE p.type = 'simple' AND p.is_active = TRUE
GROUP BY p.id, p.name
ORDER BY p.name;
