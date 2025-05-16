create table public.category (
  id         uuid primary key default uuid_generate_v4(),
  event_id   uuid not null references public.events(id) on delete cascade,
  name       text not null,
  emoji      text not null,
  is_active  boolean not null default true,
  created_at timestamptz not null default now()
);

create unique index category_event_name_ci_key
  on public.category (event_id, lower(name));

alter table public.category enable row level security;