SET search_path TO app, public;

ALTER TABLE "user" ADD COLUMN banned BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE "user" ADD COLUMN ban_reason TEXT;
ALTER TABLE "user" ADD COLUMN ban_expires TIMESTAMPTZ;

ALTER TABLE session ADD COLUMN impersonated_by TEXT;
