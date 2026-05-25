-- Load test seed. Idempotent: safe to re-run.
--
-- USAGE
--   psql "$DATABASE_URL" -v station_device_id=<uuid> -v pos_device_id=<uuid> -f seed.sql
--
-- ARGS
--   :station_device_id  UUID of an existing approved STATION device (with products
--                       assigned via device_product). Required for station scenario.
--   :pos_device_id      UUID of an existing approved POS device. Optional; falls
--                       back to NULL (POS scenario still works via binding ID).
--
-- All rows created here use loadtest_* primary keys / email prefixes for easy
-- cleanup via cleanup.sql.

\set ON_ERROR_STOP on

\if :{?station_device_id}
\else
  \set station_device_id NULL
  \echo 'WARNING: -v station_device_id=<uuid> not provided. Station scenario will return empty matches.'
\endif

\if :{?pos_device_id}
\else
  \set pos_device_id NULL
\endif

BEGIN;

-- 20 customer users + sessions
INSERT INTO "user" (id, email, name, email_verified, role)
SELECT
  'loadtest_user_customer_' || lpad(i::text, 3, '0'),
  'loadtest+customer-' || lpad(i::text, 3, '0') || '@loadtest.local',
  'Load Customer ' || i,
  TRUE,
  'customer'
FROM generate_series(1, 20) AS i
ON CONFLICT (id) DO NOTHING;

INSERT INTO session (id, user_id, token, expires_at, updated_at)
SELECT
  'loadtest_sess_customer_' || lpad(i::text, 3, '0'),
  'loadtest_user_customer_' || lpad(i::text, 3, '0'),
  'loadtest_customer_session_' || lpad(i::text, 3, '0'),
  NOW() + INTERVAL '30 days',
  NOW()
FROM generate_series(1, 20) AS i
ON CONFLICT (token) DO UPDATE SET expires_at = EXCLUDED.expires_at, updated_at = EXCLUDED.updated_at;

-- 3 admin users + sessions
INSERT INTO "user" (id, email, name, email_verified, role)
SELECT
  'loadtest_user_admin_' || lpad(i::text, 3, '0'),
  'loadtest+admin-' || lpad(i::text, 3, '0') || '@loadtest.local',
  'Load Admin ' || i,
  TRUE,
  'admin'
FROM generate_series(1, 3) AS i
ON CONFLICT (id) DO NOTHING;

INSERT INTO session (id, user_id, token, expires_at, updated_at)
SELECT
  'loadtest_sess_admin_' || lpad(i::text, 3, '0'),
  'loadtest_user_admin_' || lpad(i::text, 3, '0'),
  'loadtest_admin_session_' || lpad(i::text, 3, '0'),
  NOW() + INTERVAL '30 days',
  NOW()
FROM generate_series(1, 3) AS i
ON CONFLICT (token) DO UPDATE SET expires_at = EXCLUDED.expires_at, updated_at = EXCLUDED.updated_at;

-- 3 station device bindings.
-- token_hash = sha256(plaintext_token), matching repository.HashToken().
-- station_id points at the operator-supplied station device so device_product
-- lookups succeed during station redeem.
INSERT INTO device_binding (device_type, token_hash, name, created_by_user_id, station_id)
SELECT
  'STATION',
  encode(sha256(('loadtest_station_token_' || lpad(i::text, 3, '0'))::bytea), 'hex'),
  'Load Station ' || i,
  'loadtest_user_admin_001',
  NULLIF(:'station_device_id', 'NULL')::uuid
FROM generate_series(1, 3) AS i
ON CONFLICT (token_hash) DO UPDATE SET last_seen_at = NOW(), revoked_at = NULL, station_id = EXCLUDED.station_id;

-- 2 POS device bindings.
INSERT INTO device_binding (device_type, token_hash, name, created_by_user_id, device_id)
SELECT
  'POS',
  encode(sha256(('loadtest_pos_token_' || lpad(i::text, 3, '0'))::bytea), 'hex'),
  'Load POS ' || i,
  'loadtest_user_admin_001',
  NULLIF(:'pos_device_id', 'NULL')::uuid
FROM generate_series(1, 2) AS i
ON CONFLICT (token_hash) DO UPDATE SET last_seen_at = NOW(), revoked_at = NULL, device_id = EXCLUDED.device_id;

COMMIT;

\echo ''
\echo 'Seed summary:'
SELECT 'users'        AS kind, COUNT(*) FROM "user"          WHERE id    LIKE 'loadtest_%'
UNION ALL
SELECT 'sessions'     AS kind, COUNT(*) FROM session         WHERE token LIKE 'loadtest_%'
UNION ALL
SELECT 'station bind' AS kind, COUNT(*) FROM device_binding  WHERE name  LIKE 'Load Station%' AND revoked_at IS NULL
UNION ALL
SELECT 'pos bind'     AS kind, COUNT(*) FROM device_binding  WHERE name  LIKE 'Load POS%'     AND revoked_at IS NULL;
