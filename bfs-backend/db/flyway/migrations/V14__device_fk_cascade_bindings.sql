SET search_path TO app, public;

ALTER TABLE device_binding DROP CONSTRAINT IF EXISTS device_binding_device_id_fkey;
ALTER TABLE device_binding
    ADD CONSTRAINT device_binding_device_id_fkey
    FOREIGN KEY (device_id) REFERENCES device(id) ON DELETE CASCADE;

ALTER TABLE device_binding DROP CONSTRAINT IF EXISTS device_binding_station_id_fkey;
ALTER TABLE device_binding
    ADD CONSTRAINT device_binding_station_id_fkey
    FOREIGN KEY (station_id) REFERENCES device(id) ON DELETE CASCADE;

ALTER TABLE device_product DROP CONSTRAINT IF EXISTS device_product_device_id_fkey;
ALTER TABLE device_product
    ADD CONSTRAINT device_product_device_id_fkey
    FOREIGN KEY (device_id) REFERENCES device(id) ON DELETE RESTRICT;
