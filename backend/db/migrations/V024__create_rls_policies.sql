-- 1. tiny helpers (STABLE so PG can in-line them) -------------------
CREATE OR REPLACE FUNCTION current_app_user_id()
  RETURNS varchar(14) LANGUAGE sql STABLE AS
$$ SELECT current_setting('app.current_user_id', true)::varchar(14) $$;

CREATE OR REPLACE FUNCTION has_event_access(p_event varchar(14))
  RETURNS boolean LANGUAGE sql STABLE AS
$$ SELECT EXISTS (
       SELECT 1 FROM event_users
        WHERE event_id = p_event
          AND user_id  = current_app_user_id()
   ) $$;

CREATE OR REPLACE FUNCTION is_event_owner(p_event varchar(14))
  RETURNS boolean LANGUAGE sql STABLE AS
$$ SELECT EXISTS (
       SELECT 1 FROM events
        WHERE id = p_event
          AND owner_id = current_app_user_id()
   ) $$;

/* 2A. tables that really have user_id ------------------------------*/
DO $$
DECLARE t text;
BEGIN
  FOREACH t IN ARRAY ARRAY[
      'verification_tokens','password_reset_tokens',
      'refresh_tokens','audit_logs'
  ] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);

    EXECUTE format(
      $POL$
        DROP  POLICY IF EXISTS own_%I ON %I;
        CREATE POLICY own_%I ON %I
          FOR ALL
          USING (user_id = current_app_user_id())
          WITH  CHECK (user_id = current_app_user_id());
      $POL$, t, t, t, t);
  END LOOP;
END $$;

/* 2B. users --------------------------------------------------------*/
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
DROP  POLICY IF EXISTS self_users ON users;
CREATE POLICY self_users
  ON users FOR ALL
  USING (id = current_app_user_id())
  WITH  CHECK (id = current_app_user_id());

/* 3. event-scoped tables ------------------------------------------*/
DO $$
DECLARE t text;
BEGIN
  FOREACH t IN ARRAY ARRAY[
    'event_users','event_invites','event_devices',
    'event_categories','event_products','customer_orders'
  ] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);

    EXECUTE format(
      $EP$
        DROP  POLICY IF EXISTS event_rw_%I ON %I;
        CREATE POLICY event_rw_%I ON %I
          FOR ALL
          USING (is_event_owner(event_id) OR has_event_access(event_id))
          WITH  CHECK (is_event_owner(event_id) OR has_event_access(event_id));
      $EP$, t, t, t, t);
  END LOOP;
END $$;

/* device_pairings --------------------------------------------------*/
ALTER TABLE device_pairings ENABLE ROW LEVEL SECURITY;
DROP  POLICY IF EXISTS event_rw_device_pairings ON device_pairings;
CREATE POLICY event_rw_device_pairings
ON device_pairings FOR ALL
USING  (
         is_event_owner( (SELECT event_id FROM event_devices ed
                          WHERE ed.id = device_pairings.event_device_id) )
      OR has_event_access( (SELECT event_id FROM event_devices ed
                            WHERE ed.id = device_pairings.event_device_id) )
       )
WITH CHECK (
         is_event_owner( (SELECT event_id FROM event_devices ed
                          WHERE ed.id = device_pairings.event_device_id) )
      OR has_event_access( (SELECT event_id FROM event_devices ed
                            WHERE ed.id = device_pairings.event_device_id) )
       );

/* customer_order_items ---------------------------------------------*/
ALTER TABLE customer_order_items ENABLE ROW LEVEL SECURITY;
DROP  POLICY IF EXISTS event_rw_co_items ON customer_order_items;
CREATE POLICY event_rw_co_items
ON customer_order_items FOR ALL
USING  (
         is_event_owner( (SELECT event_id FROM customer_orders co
                          WHERE co.id = customer_order_items.customer_order_id) )
      OR has_event_access( (SELECT event_id FROM customer_orders co
                            WHERE co.id = customer_order_items.customer_order_id) )
       )
WITH CHECK (
         is_event_owner( (SELECT event_id FROM customer_orders co
                          WHERE co.id = customer_order_items.customer_order_id) )
      OR has_event_access( (SELECT event_id FROM customer_orders co
                            WHERE co.id = customer_order_items.customer_order_id) )
       );

/* 4. read-only reference tables ------------------------------------*/
ALTER TABLE roles       ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles       FORCE  ROW LEVEL SECURITY;

ALTER TABLE event_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE event_roles FORCE  ROW LEVEL SECURITY;

ALTER TABLE devices     ENABLE ROW LEVEL SECURITY;
ALTER TABLE devices     FORCE  ROW LEVEL SECURITY;

DROP  POLICY IF EXISTS read_all_roles       ON roles;
DROP  POLICY IF EXISTS read_all_event_roles ON event_roles;
DROP  POLICY IF EXISTS read_all_devices     ON devices;

CREATE POLICY read_all_roles       ON roles       FOR SELECT USING (true);
CREATE POLICY read_all_event_roles ON event_roles FOR SELECT USING (true);
CREATE POLICY read_all_devices     ON devices     FOR SELECT USING (true);