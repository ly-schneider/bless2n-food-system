--------------------------------------------------------------------
-- 1.  Create an e-mail / password user in the auth schema
--------------------------------------------------------------------
INSERT INTO auth.users (
  instance_id,
  id,
  aud,
  role,
  email,
  encrypted_password,
  email_confirmed_at,
  recovery_sent_at,
  last_sign_in_at,
  raw_app_meta_data,
  raw_user_meta_data,
  created_at,
  updated_at,
  confirmation_token,
  email_change,
  email_change_token_new,
  recovery_token
)
VALUES (
  '00000000-0000-0000-0000-000000000000',
  uuid_generate_v4(),                 -- primary key
  'authenticated',
  'authenticated',
  'levyn.schneider@leys.ch',
  crypt('Test1234!', gen_salt('bf')),  -- bcrypt
  current_timestamp,
  current_timestamp,
  current_timestamp,
  '{"provider":"email","providers":["email"]}',
  '{}'::jsonb,
  current_timestamp,
  current_timestamp,
  '', '', '', ''
);

--------------------------------------------------------------------
-- 2.  Link the “email” identity row
--------------------------------------------------------------------
INSERT INTO auth.identities (
  id,
  user_id,
  identity_data,
  provider,
  provider_id,
  last_sign_in_at,
  created_at,
  updated_at
)
VALUES (
  uuid_generate_v4(),
  (SELECT id FROM auth.users WHERE email = 'levyn.schneider@leys.ch'),
  format(
    '{"sub":"%s","email":"%s"}',
    (SELECT id FROM auth.users WHERE email = 'levyn.schneider@leys.ch')::text,
    'levyn.schneider@leys.ch'
  )::jsonb,
  'email',
  uuid_generate_v4(),                 -- provider_id is arbitrary for e-mail
  current_timestamp,
  current_timestamp,
  current_timestamp
);

--------------------------------------------------------------------
-- 3.  Insert a matching profile row in public.users
--------------------------------------------------------------------
WITH src AS (
  SELECT id, email, created_at
  FROM   auth.users
  WHERE  email = 'levyn.schneider@leys.ch'
)
INSERT INTO public.users (
  id,
  firstname,
  lastname,
  email,
  is_site_admin,
  is_disabled,
  disabled_reason,
  created_at
)
SELECT
  id,
  'Levyn',               -- firstname
  'Schneider',           -- lastname
  email,
  true,                  -- is_site_admin
  false,                 -- is_disabled
  NULL,                  -- disabled_reason
  created_at
FROM src;

--------------------------------------------------------------------
-- 4.  Insert devices
--------------------------------------------------------------------
INSERT INTO public.devices (
  id,
  serial_number
  model,
  is_active,
  registered_at
)
VALUES (
  uuid_generate_v4(),
  'SERIAL_1',
  'Samsung Galaxy Tab A9+',
  true,
  current_timestamp
), (
  uuid_generate_v4(),
  'SERIAL_2',
  'Samsung Galaxy Tab A9+',
  true,
  current_timestamp
), (
  uuid_generate_v4(),
  'SERIAL_3',
  'Samsung Galaxy Tab A9+',
  true,
  current_timestamp
), (
  uuid_generate_v4(),
  'SERIAL_4',
  'Samsung Galaxy Tab A9+',
  true,
  current_timestamp
), (
  uuid_generate_v4(),
  'SERIAL_5',
  'Samsung Galaxy Tab A9+',
  true,
  current_timestamp
), (
  uuid_generate_v4(),
  'SERIAL_6',
  'Samsung Galaxy Tab A9+',
  true,
  current_timestamp
), (
  uuid_generate_v4(),
  'SERIAL_7',
  'Samsung Galaxy Tab A9+',
  true,
  current_timestamp
), (
  uuid_generate_v4(),
  'SERIAL_8',
  'Samsung Galaxy Tab A9+',
  true,
  current_timestamp
), (
  uuid_generate_v4(),
  'SERIAL_9',
  'Samsung Galaxy Tab A9+',
  true,
  current_timestamp
), (
  uuid_generate_v4(),
  'SERIAL_10',
  'Samsung Galaxy Tab A9+',
  true,
  current_timestamp
);

--------------------------------------------------------------------
-- 5.  Insert event user roles
--------------------------------------------------------------------
INSERT INTO public.event_user_roles (
  id,
  name,
  display_name,
  description,
  is_default,
  is_active,
  created_at
)
VALUES (
  uuid_generate_v4(),
  'admin',
  'Administrator',
  'Event Administrator mit vollen Rechten.',
  true,
  true,
  current_timestamp
), (
  uuid_generate_v4(),
  'bearbeiter',
  'Bearbeiter',
  'Event Bearbeiter mit eingeschränkten Rechten.',
  false,
  true,
  current_timestamp
), (
  uuid_generate_v4(),
  'gast',
  'Gast',
  'Event Gast mit nur Lesezugriff auf das Event.',
  false,
  true,
  current_timestamp
);