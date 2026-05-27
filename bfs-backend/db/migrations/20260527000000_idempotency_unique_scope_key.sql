DELETE FROM idempotency a
USING idempotency b
WHERE a.scope = b.scope
  AND a.key = b.key
  AND (a.expires_at, a.id) < (b.expires_at, b.id);

DROP INDEX IF EXISTS idx_idempotency_scope_key;

CREATE UNIQUE INDEX idempotency_scope_key_unique ON idempotency (scope, key);
