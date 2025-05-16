create table public.order_items (
  id              uuid primary key default uuid_generate_v4(),
  order_id        uuid not null references public.orders(id) on delete cascade,
  product_id      uuid not null references public.products(id),
  quantity        integer not null check (quantity > 0),
  price_per_unit  numeric(12,2) not null check (price_per_unit >= 0),
  unique (order_id, product_id)
);
alter table public.order_items enable row level security;