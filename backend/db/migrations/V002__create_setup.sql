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

CREATE DOMAIN argon2_hash AS varchar(120)   
  COLLATE "C"   
  CHECK (     
    VALUE ~ '^\$argon2id\$v=\d+\$m=\d+,t=\d+,p=\d+\$[A-Za-z0-9+/]+\$[A-Za-z0-9+/]+$' -- Argon2id hash format
  );

CREATE TYPE device_pairing_role_enum AS ENUM ('customer', 'cashier');

CREATE TYPE order_status_enum AS ENUM ('pending', 'completed', 'cancelled');