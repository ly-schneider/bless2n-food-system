create table public.event_invites (
  id            uuid primary key default uuid_generate_v4(),
  event_id      uuid not null references public.events(id) on delete cascade,
  role_id       uuid not null references public.event_user_roles(id),
  invitee_email text not null,
  expires_at    timestamptz not null
);
alter table public.event_invites enable row level security;