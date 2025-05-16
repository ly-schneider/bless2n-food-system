create table public.users (
  id              uuid REFERENCES auth.users ON DELETE CASCADE PRIMARY KEY,
  firstname       text        not null,
  lastname        text        not null,
  email           text        not null unique,
  is_site_admin   boolean     not null default false,
  is_disabled     boolean     not null default false,
  disabled_reason text,
  created_at      timestamptz not null default now()
);
alter table public.users enable row level security;