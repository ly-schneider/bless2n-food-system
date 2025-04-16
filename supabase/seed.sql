INSERT INTO auth.users ( instance_id, id, aud, role, email, encrypted_password, email_confirmed_at, recovery_sent_at, last_sign_in_at, raw_app_meta_data, raw_user_meta_data, created_at, updated_at, confirmation_token, email_change, email_change_token_new, recovery_token) 
VALUES 
  ('00000000-0000-0000-0000-000000000000', uuid_generate_v4(), 'authenticated', 'authenticated', 'bar@icf-bern.ch', crypt('B@rJoyride2025!', gen_salt('bf')), current_timestamp, current_timestamp, current_timestamp, '{"provider":"email","providers":["email"]}', '{}', current_timestamp, current_timestamp, '', '', '', '');

INSERT INTO auth.identities (id, user_id, identity_data, provider, provider_id, last_sign_in_at, created_at, updated_at)
VALUES 
  (uuid_generate_v4(), (SELECT id FROM auth.users WHERE email = 'bar@icf-bern.ch'), format('{"sub":"%s","email":"%s"}', (SELECT id FROM auth.users WHERE email = 'bar@icf-bern.ch')::text, 'bar@icf-bern.ch')::jsonb, 'email', uuid_generate_v4(), current_timestamp, current_timestamp, current_timestamp);

INSERT INTO public.admins (id, email, created_at, updated_at)
SELECT 
  id,
  email,
  created_at,
  updated_at
FROM auth.users;

INSERT INTO public.categories (name)
VALUES 
  ('Drinks'),
  ('Foods'),
  ('Sweets');

INSERT INTO public.products (name, price, thumbnail_url, available, created_at, updated_at, created_by, category_id)
VALUES 
  ('Prosecco Maccari Extra Dry', 7.00, NULL, TRUE, NOW(), NOW(), (SELECT id FROM public.admins WHERE email = 'bar@icf-bern.ch'), (SELECT id FROM public.categories WHERE name = 'Drinks')),
  ('Joyride (alkoholfrei)', 5.00, NULL, TRUE, NOW(), NOW(), (SELECT id FROM public.admins WHERE email = 'bar@icf-bern.ch'), (SELECT id FROM public.categories WHERE name = 'Drinks')),
  ('Jolys Drive', 7, NULL, TRUE, NOW(), NOW(), (SELECT id FROM public.admins WHERE email = 'bar@icf-bern.ch'), (SELECT id FROM public.categories WHERE name = 'Drinks')),
  ('Bärner Müntschi', 4.50, NULL, TRUE, NOW(), NOW(), (SELECT id FROM public.admins WHERE email = 'bar@icf-bern.ch'), (SELECT id FROM public.categories WHERE name = 'Drinks')),
  ('Mineralwasser und Süssgetränke', 4.00, NULL, TRUE, NOW(), NOW(), (SELECT id FROM public.admins WHERE email = 'bar@icf-bern.ch'), (SELECT id FROM public.categories WHERE name = 'Drinks')),
  ('Kaffee / Espresso / Tee', 4.00, NULL, TRUE, NOW(), NOW(), (SELECT id FROM public.admins WHERE email = 'bar@icf-bern.ch'), (SELECT id FROM public.categories WHERE name = 'Drinks')),
  
  ('Hiobs Hot Dog', 7.00, NULL, TRUE, NOW(), NOW(), (SELECT id FROM public.admins WHERE email = 'bar@icf-bern.ch'), (SELECT id FROM public.categories WHERE name = 'Foods')),
  ('Joys Hot Dog', 7.00, NULL, TRUE, NOW(), NOW(), (SELECT id FROM public.admins WHERE email = 'bar@icf-bern.ch'), (SELECT id FROM public.categories WHERE name = 'Foods')),
  ('Zweifel Chips', 2.50, NULL, TRUE, NOW(), NOW(), (SELECT id FROM public.admins WHERE email = 'bar@icf-bern.ch'), (SELECT id FROM public.categories WHERE name = 'Foods')),
  
  ('Waffel mit Engelszucker', 4.00, NULL, TRUE, NOW(), NOW(), (SELECT id FROM public.admins WHERE email = 'bar@icf-bern.ch'), (SELECT id FROM public.categories WHERE name = 'Sweets')),
  ('Waffel mit frischen Erdbeeren und Schlagrahm', 7.00, NULL, TRUE, NOW(), NOW(), (SELECT id FROM public.admins WHERE email = 'bar@icf-bern.ch'), (SELECT id FROM public.categories WHERE name = 'Sweets')),
  ('Waffel mit Kokoscreme und Ananas', 7.00, NULL, TRUE, NOW(), NOW(), (SELECT id FROM public.admins WHERE email = 'bar@icf-bern.ch'), (SELECT id FROM public.categories WHERE name = 'Sweets'));
