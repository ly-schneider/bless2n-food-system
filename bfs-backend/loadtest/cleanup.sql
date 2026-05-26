-- Remove all rows seeded by loadtest/seed.sql and loadtest/stock.sql.
--
-- USAGE
--   psql "$DATABASE_URL" -f cleanup.sql

\set ON_ERROR_STOP on

BEGIN;

-- order.customer_id is ON DELETE SET NULL, so user-deletion alone leaves
-- orders behind with null attribution. Delete them explicitly by contact email.
DELETE FROM "order"           WHERE contact_email LIKE 'loadtest+%@loadtest.local';
DELETE FROM inventory_ledger  WHERE created_by = 'loadtest_user_admin_001' AND reason = 'manual_adjust';
DELETE FROM device_binding    WHERE name LIKE 'Load Station%' OR name LIKE 'Load POS%';
DELETE FROM session           WHERE token LIKE 'loadtest_%';
DELETE FROM "user"            WHERE id LIKE 'loadtest_%';

COMMIT;

\echo 'Loadtest data removed.'
