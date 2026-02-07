SET search_path TO app, public;

ALTER TABLE "user" ADD COLUMN is_club_100 BOOLEAN NOT NULL DEFAULT FALSE;
