-- Load test seed. Idempotent: safe to re-run.
--
-- USAGE
--   psql "$DATABASE_URL" -v station_device_id=<uuid> -v pos_device_id=<uuid> -f seed.sql
--
-- station_device_id is required for the station scenario (it must be an
-- approved STATION device with products assigned via device_product).
-- pos_device_id is optional; without it the POS device_binding's device_id
-- stays null and order_payment FK can't resolve (cash payments will fail).

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

-- token_hash = sha256(plaintext_token), matching repository.HashToken().
INSERT INTO device_binding (device_type, token_hash, name, created_by_user_id, station_id)
SELECT
  'STATION',
  encode(sha256(('loadtest_station_token_' || lpad(i::text, 3, '0'))::bytea), 'hex'),
  'Load Station ' || i,
  'loadtest_user_admin_001',
  NULLIF(:'station_device_id', 'NULL')::uuid
FROM generate_series(1, 3) AS i
ON CONFLICT (token_hash) DO UPDATE SET last_seen_at = NOW(), revoked_at = NULL, station_id = EXCLUDED.station_id;

INSERT INTO device_binding (device_type, token_hash, name, created_by_user_id, device_id)
SELECT
  'POS',
  encode(sha256(('loadtest_pos_token_' || lpad(i::text, 3, '0'))::bytea), 'hex'),
  'Load POS ' || i,
  'loadtest_user_admin_001',
  NULLIF(:'pos_device_id', 'NULL')::uuid
FROM generate_series(1, 2) AS i
ON CONFLICT (token_hash) DO UPDATE SET last_seen_at = NOW(), revoked_at = NULL, device_id = EXCLUDED.device_id;

-- Device auth middleware looks up device_binding by token_hash AND session by
-- raw token; without the session row every device request returns 401.
INSERT INTO session (id, user_id, token, expires_at, updated_at)
SELECT
  'loadtest_sess_station_' || lpad(i::text, 3, '0'),
  'loadtest_user_admin_001',
  'loadtest_station_token_' || lpad(i::text, 3, '0'),
  NOW() + INTERVAL '30 days',
  NOW()
FROM generate_series(1, 3) AS i
ON CONFLICT (token) DO UPDATE SET expires_at = EXCLUDED.expires_at, updated_at = EXCLUDED.updated_at;

INSERT INTO session (id, user_id, token, expires_at, updated_at)
SELECT
  'loadtest_sess_pos_' || lpad(i::text, 3, '0'),
  'loadtest_user_admin_001',
  'loadtest_pos_token_' || lpad(i::text, 3, '0'),
  NOW() + INTERVAL '30 days',
  NOW()
FROM generate_series(1, 2) AS i
ON CONFLICT (token) DO UPDATE SET expires_at = EXCLUDED.expires_at, updated_at = EXCLUDED.updated_at;

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
