create table public.event_device_pairings (
  id                 uuid primary key default uuid_generate_v4(),
  customer_device_id uuid unique references public.event_devices(id) on delete cascade,
  cashier_device_id  uuid unique references public.event_devices(id) on delete cascade,
  linked_at          timestamptz not null
);
alter table public.event_device_pairings enable row level security;