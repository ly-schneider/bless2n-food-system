/* ─────────────────────────  EXTENSIONS  ───────────────────────── */
CREATE EXTENSION IF NOT EXISTS pgcrypto;      -- for hashing / crypto funcs
CREATE EXTENSION IF NOT EXISTS pgaudit;       -- detailed audit logging

/* ───────────────────────  DOMAIN / ENUMS  ─────────────────────── */
CREATE DOMAIN nano_id AS varchar(14) 
    COLLATE "C"
    CHECK (length(VALUE) = 14);