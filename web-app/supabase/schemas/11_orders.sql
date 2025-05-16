create type order_status as enum ('pending','completed','cancelled');
create table public.orders (
  id          uuid primary key default uuid_generate_v4(),
  event_id    uuid not null references public.events(id) on delete cascade,
  device_id   uuid not null references public.devices(id) on delete set null,
  total       numeric(12,2) not null check (total >= 0),
  status      order_status not null,
  ordered_at  timestamptz  not null default now()
);
alter table public.orders enable row level security;