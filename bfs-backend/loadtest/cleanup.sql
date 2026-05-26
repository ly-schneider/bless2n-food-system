-- Remove all rows seeded by loadtest/seed.sql.
-- Run after a load test to leave staging clean.
--
-- USAGE
--   psql "$DATABASE_URL" -f cleanup.sql

\set ON_ERROR_STOP on

BEGIN;

-- Orders created by loadtest users get cascade-removed via FK on customer_id
-- (ON DELETE SET NULL), so they stay but lose attribution. Remove them
-- explicitly if you want a fully clean slate.
DELETE FROM "order"           WHERE contact_email LIKE 'loadtest+%@loadtest.local';
DELETE FROM inventory_ledger  WHERE created_by = 'loadtest_user_admin_001' AND reason = 'manual_adjust';
DELETE FROM device_binding    WHERE name LIKE 'Load Station%' OR name LIKE 'Load POS%';
DELETE FROM session           WHERE token LIKE 'loadtest_%';
DELETE FROM "user"            WHERE id LIKE 'loadtest_%';

COMMIT;

\echo 'Loadtest data removed.'
