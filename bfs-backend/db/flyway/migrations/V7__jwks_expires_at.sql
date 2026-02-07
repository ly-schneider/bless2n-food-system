SET search_path TO app, public;

ALTER TABLE jwks ADD COLUMN expires_at TIMESTAMPTZ;
