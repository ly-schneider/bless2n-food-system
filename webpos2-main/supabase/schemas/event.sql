CREATE TABLE public.event (
    event_id uuid DEFAULT uuid_generate_v4() NOT NULL,
    name varchar(255) NOT NULL,
    time_from timestamptz NOT NULL,
    time_to timestamptz NOT NULL,
    created_at timestamptz DEFAULT now(),
    location_id uuid
);