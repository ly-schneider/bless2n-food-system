create table public.products (
  id          uuid primary key default uuid_generate_v4(),
  event_id    uuid not null references public.events(id) on delete cascade,
  category_id uuid references public.category(id),
  name        text not null,
  emoji       text not null,
  price       numeric(12,2) check (price >= 0),
  is_active   boolean not null default true,
  created_at  timestamptz not null default now()
);

create unique index products_event_name_ci_key
  on public.products (event_id, lower(name));

alter table public.products enable row level security;