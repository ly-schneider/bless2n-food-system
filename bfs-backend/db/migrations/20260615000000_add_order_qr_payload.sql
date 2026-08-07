-- Signed Ed25519 pickup token (see internal/qrsign), persisted at creation.
-- The signing key lives in env (QR_ED25519_PRIVATE_SEED), not the database.

ALTER TABLE "order"
  ADD COLUMN qr_payload TEXT;
