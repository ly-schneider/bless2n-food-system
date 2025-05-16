create table public.event_user_roles (
  id           uuid primary key default uuid_generate_v4(),
  name         text not null unique check (name = lower(name)),
  display_name text not null,
  description  text,
  is_default   boolean not null default false,
  is_active    boolean not null default true,
  created_at   timestamptz not null default now()
);
alter table public.event_user_roles enable row level security;