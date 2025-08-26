CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END $$;

CREATE TRIGGER trg_set_updated_at_users
BEFORE UPDATE ON users
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_events
BEFORE UPDATE ON events
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_devices
BEFORE UPDATE ON devices
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_device_pairings
BEFORE UPDATE ON device_pairings
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_event_categories
BEFORE UPDATE ON event_categories
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_set_updated_at_event_products
BEFORE UPDATE ON event_products
FOR EACH ROW
WHEN (OLD IS DISTINCT FROM NEW)
EXECUTE FUNCTION set_updated_at();