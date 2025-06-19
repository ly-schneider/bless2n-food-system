/* ─────────────────────────  EXTENSIONS  ───────────────────────── */
CREATE EXTENSION IF NOT EXISTS pgcrypto;      -- for hashing / crypto funcs
CREATE EXTENSION IF NOT EXISTS pgaudit;       -- detailed audit logging

/* ───────────────────────  DOMAIN / ENUMS  ─────────────────────── */

/*
"C" collation ⇒ raw byte-wise (POSIX) ordering, fastest & locale-neutral
Ideal for hashes/IDs; keeps index results stable across servers/locale upgrades
*/

CREATE DOMAIN nano_id AS varchar(14)
  COLLATE "C"
  CHECK (
    octet_length(VALUE) = 14                               -- exactly 14 bytes
    AND VALUE ~ '^[A-Za-z0-9]{14}$'                        -- Nano alphabet
  );

CREATE DOMAIN bcrypt_hash AS varchar(60)
  COLLATE "C"
  CHECK (
    octet_length(VALUE) = 60                               -- exactly 60 bytes
    AND VALUE ~ '^\$2[abxy]?\$\d{2}\$[./A-Za-z0-9]{53}$'   -- bcrypt format
  );

CREATE DOMAIN sha256_hash AS varchar(64)
  COLLATE "C"
  CHECK (
        octet_length(VALUE) = 64                           -- exactly 64 bytes
    AND VALUE ~ '^[0-9a-f]{64}$'                           -- lowercase hex only
  );

CREATE TYPE device_pairing_role_enum AS ENUM ('customer', 'cashier');

CREATE TYPE order_status_enum AS ENUM ('pending', 'completed', 'cancelled');