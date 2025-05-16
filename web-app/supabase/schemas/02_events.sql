create table public.events (
  id          uuid primary key default uuid_generate_v4(),
  owner_id    uuid not null references public.users(id) on delete cascade,
  name        text not null,
  start_date  date not null,
  end_date    date not null check (end_date >= start_date),
  created_at  timestamptz not null default now()
);
alter table public.events enable row level security;