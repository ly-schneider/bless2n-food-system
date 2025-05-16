create table public.event_devices (
  id         uuid primary key default uuid_generate_v4(),
  event_id   uuid not null references public.events(id)  on delete cascade,
  device_id  uuid not null references public.devices(id) on delete cascade,
  assigned_at timestamptz not null default now(),
  unique (event_id, device_id)
);
alter table public.event_devices enable row level security;