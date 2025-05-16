create table public.event_users (
  id         uuid primary key default uuid_generate_v4(),
  event_id   uuid not null references public.events(id)  on delete cascade,
  user_id    uuid not null references public.users(id)   on delete cascade,
  role_id    uuid not null references public.event_user_roles(id),
  invited_at timestamptz not null default now(),
  joined_at  timestamptz not null,
  unique (event_id, user_id)   -- one row per participant per event
);
alter table public.event_users enable row level security;