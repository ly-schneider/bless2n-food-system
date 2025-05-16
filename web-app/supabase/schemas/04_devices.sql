create table public.devices (
  id             uuid primary key default uuid_generate_v4(),
  serial_number  text not null unique,
  model          text not null,
  is_active      boolean not null default true,
  registered_at  timestamptz not null default now()
);
alter table public.devices enable row level security;