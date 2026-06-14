-- Migrate entity ids from uuidv7 to app-generated nanoid(12).
--
-- Forward-only: columns are widened to VARCHAR(36) preserving existing uuid
-- values (uuid::text); new rows get nanoids from the app, old rows keep their
-- uuids. Foreign keys must be dropped before the referenced/referencing column
-- types can change, then recreated with their original ON DELETE behaviour.

-- 1. Drop foreign-key constraints

ALTER TABLE product DROP CONSTRAINT IF EXISTS product_category_id_fkey;
ALTER TABLE product DROP CONSTRAINT IF EXISTS product_jeton_id_fkey;
ALTER TABLE menu_slot DROP CONSTRAINT IF EXISTS menu_slot_menu_product_id_fkey;
ALTER TABLE menu_slot_option DROP CONSTRAINT IF EXISTS menu_slot_option_menu_slot_id_fkey;
ALTER TABLE menu_slot_option DROP CONSTRAINT IF EXISTS menu_slot_option_option_product_id_fkey;
ALTER TABLE device_product DROP CONSTRAINT IF EXISTS device_product_device_id_fkey;
ALTER TABLE device_product DROP CONSTRAINT IF EXISTS device_product_product_id_fkey;
ALTER TABLE device_binding DROP CONSTRAINT IF EXISTS device_binding_station_id_fkey;
ALTER TABLE device_binding DROP CONSTRAINT IF EXISTS device_binding_device_id_fkey;
ALTER TABLE order_payment DROP CONSTRAINT IF EXISTS order_payment_order_id_fkey;
ALTER TABLE order_payment DROP CONSTRAINT IF EXISTS order_payment_device_id_fkey;
ALTER TABLE order_line DROP CONSTRAINT IF EXISTS order_line_order_id_fkey;
ALTER TABLE order_line DROP CONSTRAINT IF EXISTS order_line_product_id_fkey;
ALTER TABLE order_line DROP CONSTRAINT IF EXISTS order_line_parent_line_id_fkey;
ALTER TABLE order_line DROP CONSTRAINT IF EXISTS order_line_menu_slot_id_fkey;
ALTER TABLE order_line_redemption DROP CONSTRAINT IF EXISTS order_line_redemption_order_line_id_fkey;
ALTER TABLE inventory_ledger DROP CONSTRAINT IF EXISTS inventory_ledger_product_id_fkey;
ALTER TABLE inventory_ledger DROP CONSTRAINT IF EXISTS inventory_ledger_order_id_fkey;
ALTER TABLE inventory_ledger DROP CONSTRAINT IF EXISTS inventory_ledger_order_line_id_fkey;
ALTER TABLE inventory_ledger DROP CONSTRAINT IF EXISTS inventory_ledger_device_id_fkey;
ALTER TABLE club100_redemption DROP CONSTRAINT IF EXISTS club100_redemption_order_id_fkey;
ALTER TABLE club100_free_product DROP CONSTRAINT IF EXISTS club100_free_product_product_id_fkey;
ALTER TABLE volunteer_campaign_product DROP CONSTRAINT IF EXISTS volunteer_campaign_product_campaign_id_fkey;
ALTER TABLE volunteer_campaign_product DROP CONSTRAINT IF EXISTS volunteer_campaign_product_product_id_fkey;
ALTER TABLE volunteer_redemption DROP CONSTRAINT IF EXISTS volunteer_redemption_campaign_id_fkey;
ALTER TABLE volunteer_redemption DROP CONSTRAINT IF EXISTS volunteer_redemption_order_id_fkey;
ALTER TABLE volunteer_redemption DROP CONSTRAINT IF EXISTS volunteer_redemption_station_device_id_fkey;

-- 2. Drop uuidv7() column defaults

ALTER TABLE category ALTER COLUMN id DROP DEFAULT;
ALTER TABLE jeton ALTER COLUMN id DROP DEFAULT;
ALTER TABLE product ALTER COLUMN id DROP DEFAULT;
ALTER TABLE menu_slot ALTER COLUMN id DROP DEFAULT;
ALTER TABLE device ALTER COLUMN id DROP DEFAULT;
ALTER TABLE device_binding ALTER COLUMN id DROP DEFAULT;
ALTER TABLE "order" ALTER COLUMN id DROP DEFAULT;
ALTER TABLE order_payment ALTER COLUMN id DROP DEFAULT;
ALTER TABLE order_line ALTER COLUMN id DROP DEFAULT;
ALTER TABLE order_line_redemption ALTER COLUMN id DROP DEFAULT;
ALTER TABLE inventory_ledger ALTER COLUMN id DROP DEFAULT;
ALTER TABLE idempotency ALTER COLUMN id DROP DEFAULT;
ALTER TABLE admin_invite ALTER COLUMN id DROP DEFAULT;
ALTER TABLE club100_redemption ALTER COLUMN id DROP DEFAULT;
ALTER TABLE volunteer_campaign ALTER COLUMN id DROP DEFAULT;
ALTER TABLE volunteer_campaign ALTER COLUMN claim_token DROP DEFAULT;
ALTER TABLE volunteer_redemption ALTER COLUMN id DROP DEFAULT;

-- 3. Widen UUID columns to VARCHAR(36), preserving existing values

ALTER TABLE category ALTER COLUMN id TYPE VARCHAR(36) USING id::text;
ALTER TABLE jeton ALTER COLUMN id TYPE VARCHAR(36) USING id::text;

ALTER TABLE product ALTER COLUMN id TYPE VARCHAR(36) USING id::text;
ALTER TABLE product ALTER COLUMN category_id TYPE VARCHAR(36) USING category_id::text;
ALTER TABLE product ALTER COLUMN jeton_id TYPE VARCHAR(36) USING jeton_id::text;

ALTER TABLE menu_slot ALTER COLUMN id TYPE VARCHAR(36) USING id::text;
ALTER TABLE menu_slot ALTER COLUMN menu_product_id TYPE VARCHAR(36) USING menu_product_id::text;

ALTER TABLE menu_slot_option ALTER COLUMN menu_slot_id TYPE VARCHAR(36) USING menu_slot_id::text;
ALTER TABLE menu_slot_option ALTER COLUMN option_product_id TYPE VARCHAR(36) USING option_product_id::text;

ALTER TABLE device ALTER COLUMN id TYPE VARCHAR(36) USING id::text;

ALTER TABLE device_product ALTER COLUMN device_id TYPE VARCHAR(36) USING device_id::text;
ALTER TABLE device_product ALTER COLUMN product_id TYPE VARCHAR(36) USING product_id::text;

ALTER TABLE device_binding ALTER COLUMN id TYPE VARCHAR(36) USING id::text;
ALTER TABLE device_binding ALTER COLUMN station_id TYPE VARCHAR(36) USING station_id::text;
ALTER TABLE device_binding ALTER COLUMN device_id TYPE VARCHAR(36) USING device_id::text;

ALTER TABLE "order" ALTER COLUMN id TYPE VARCHAR(36) USING id::text;

ALTER TABLE order_payment ALTER COLUMN id TYPE VARCHAR(36) USING id::text;
ALTER TABLE order_payment ALTER COLUMN order_id TYPE VARCHAR(36) USING order_id::text;
ALTER TABLE order_payment ALTER COLUMN device_id TYPE VARCHAR(36) USING device_id::text;

ALTER TABLE order_line ALTER COLUMN id TYPE VARCHAR(36) USING id::text;
ALTER TABLE order_line ALTER COLUMN order_id TYPE VARCHAR(36) USING order_id::text;
ALTER TABLE order_line ALTER COLUMN product_id TYPE VARCHAR(36) USING product_id::text;
ALTER TABLE order_line ALTER COLUMN parent_line_id TYPE VARCHAR(36) USING parent_line_id::text;
ALTER TABLE order_line ALTER COLUMN menu_slot_id TYPE VARCHAR(36) USING menu_slot_id::text;

ALTER TABLE order_line_redemption ALTER COLUMN id TYPE VARCHAR(36) USING id::text;
ALTER TABLE order_line_redemption ALTER COLUMN order_line_id TYPE VARCHAR(36) USING order_line_id::text;

ALTER TABLE inventory_ledger ALTER COLUMN id TYPE VARCHAR(36) USING id::text;
ALTER TABLE inventory_ledger ALTER COLUMN product_id TYPE VARCHAR(36) USING product_id::text;
ALTER TABLE inventory_ledger ALTER COLUMN order_id TYPE VARCHAR(36) USING order_id::text;
ALTER TABLE inventory_ledger ALTER COLUMN order_line_id TYPE VARCHAR(36) USING order_line_id::text;
ALTER TABLE inventory_ledger ALTER COLUMN device_id TYPE VARCHAR(36) USING device_id::text;

ALTER TABLE idempotency ALTER COLUMN id TYPE VARCHAR(36) USING id::text;

ALTER TABLE admin_invite ALTER COLUMN id TYPE VARCHAR(36) USING id::text;

ALTER TABLE club100_redemption ALTER COLUMN id TYPE VARCHAR(36) USING id::text;
ALTER TABLE club100_redemption ALTER COLUMN order_id TYPE VARCHAR(36) USING order_id::text;

ALTER TABLE club100_free_product ALTER COLUMN product_id TYPE VARCHAR(36) USING product_id::text;

ALTER TABLE volunteer_campaign ALTER COLUMN id TYPE VARCHAR(36) USING id::text;
ALTER TABLE volunteer_campaign ALTER COLUMN claim_token TYPE VARCHAR(36) USING claim_token::text;

ALTER TABLE volunteer_campaign_product ALTER COLUMN campaign_id TYPE VARCHAR(36) USING campaign_id::text;
ALTER TABLE volunteer_campaign_product ALTER COLUMN product_id TYPE VARCHAR(36) USING product_id::text;

ALTER TABLE volunteer_redemption ALTER COLUMN id TYPE VARCHAR(36) USING id::text;
ALTER TABLE volunteer_redemption ALTER COLUMN campaign_id TYPE VARCHAR(36) USING campaign_id::text;
ALTER TABLE volunteer_redemption ALTER COLUMN order_id TYPE VARCHAR(36) USING order_id::text;
ALTER TABLE volunteer_redemption ALTER COLUMN station_device_id TYPE VARCHAR(36) USING station_device_id::text;

-- 4. Recreate foreign-key constraints

ALTER TABLE product
    ADD CONSTRAINT product_category_id_fkey FOREIGN KEY (category_id) REFERENCES category (id) ON DELETE RESTRICT,
    ADD CONSTRAINT product_jeton_id_fkey FOREIGN KEY (jeton_id) REFERENCES jeton (id) ON DELETE SET NULL;

ALTER TABLE menu_slot
    ADD CONSTRAINT menu_slot_menu_product_id_fkey FOREIGN KEY (menu_product_id) REFERENCES product (id) ON DELETE CASCADE;

ALTER TABLE menu_slot_option
    ADD CONSTRAINT menu_slot_option_menu_slot_id_fkey FOREIGN KEY (menu_slot_id) REFERENCES menu_slot (id) ON DELETE CASCADE,
    ADD CONSTRAINT menu_slot_option_option_product_id_fkey FOREIGN KEY (option_product_id) REFERENCES product (id) ON DELETE RESTRICT;

ALTER TABLE device_product
    ADD CONSTRAINT device_product_device_id_fkey FOREIGN KEY (device_id) REFERENCES device (id) ON DELETE RESTRICT,
    ADD CONSTRAINT device_product_product_id_fkey FOREIGN KEY (product_id) REFERENCES product (id) ON DELETE CASCADE;

ALTER TABLE device_binding
    ADD CONSTRAINT device_binding_station_id_fkey FOREIGN KEY (station_id) REFERENCES device (id) ON DELETE CASCADE,
    ADD CONSTRAINT device_binding_device_id_fkey FOREIGN KEY (device_id) REFERENCES device (id) ON DELETE CASCADE;

ALTER TABLE order_payment
    ADD CONSTRAINT order_payment_order_id_fkey FOREIGN KEY (order_id) REFERENCES "order" (id) ON DELETE CASCADE,
    ADD CONSTRAINT order_payment_device_id_fkey FOREIGN KEY (device_id) REFERENCES device (id) ON DELETE SET NULL;

ALTER TABLE order_line
    ADD CONSTRAINT order_line_order_id_fkey FOREIGN KEY (order_id) REFERENCES "order" (id) ON DELETE CASCADE,
    ADD CONSTRAINT order_line_product_id_fkey FOREIGN KEY (product_id) REFERENCES product (id) ON DELETE RESTRICT,
    ADD CONSTRAINT order_line_parent_line_id_fkey FOREIGN KEY (parent_line_id) REFERENCES order_line (id) ON DELETE CASCADE,
    ADD CONSTRAINT order_line_menu_slot_id_fkey FOREIGN KEY (menu_slot_id) REFERENCES menu_slot (id) ON DELETE SET NULL;

ALTER TABLE order_line_redemption
    ADD CONSTRAINT order_line_redemption_order_line_id_fkey FOREIGN KEY (order_line_id) REFERENCES order_line (id) ON DELETE CASCADE;

ALTER TABLE inventory_ledger
    ADD CONSTRAINT inventory_ledger_product_id_fkey FOREIGN KEY (product_id) REFERENCES product (id) ON DELETE CASCADE,
    ADD CONSTRAINT inventory_ledger_order_id_fkey FOREIGN KEY (order_id) REFERENCES "order" (id) ON DELETE SET NULL,
    ADD CONSTRAINT inventory_ledger_order_line_id_fkey FOREIGN KEY (order_line_id) REFERENCES order_line (id) ON DELETE SET NULL,
    ADD CONSTRAINT inventory_ledger_device_id_fkey FOREIGN KEY (device_id) REFERENCES device (id) ON DELETE SET NULL;

ALTER TABLE club100_redemption
    ADD CONSTRAINT club100_redemption_order_id_fkey FOREIGN KEY (order_id) REFERENCES "order" (id) ON DELETE CASCADE;

ALTER TABLE club100_free_product
    ADD CONSTRAINT club100_free_product_product_id_fkey FOREIGN KEY (product_id) REFERENCES product (id) ON DELETE CASCADE;

ALTER TABLE volunteer_campaign_product
    ADD CONSTRAINT volunteer_campaign_product_campaign_id_fkey FOREIGN KEY (campaign_id) REFERENCES volunteer_campaign (id) ON DELETE CASCADE,
    ADD CONSTRAINT volunteer_campaign_product_product_id_fkey FOREIGN KEY (product_id) REFERENCES product (id) ON DELETE RESTRICT;

ALTER TABLE volunteer_redemption
    ADD CONSTRAINT volunteer_redemption_campaign_id_fkey FOREIGN KEY (campaign_id) REFERENCES volunteer_campaign (id) ON DELETE CASCADE,
    ADD CONSTRAINT volunteer_redemption_order_id_fkey FOREIGN KEY (order_id) REFERENCES "order" (id) ON DELETE RESTRICT,
    ADD CONSTRAINT volunteer_redemption_station_device_id_fkey FOREIGN KEY (station_device_id) REFERENCES device (id) ON DELETE SET NULL;
