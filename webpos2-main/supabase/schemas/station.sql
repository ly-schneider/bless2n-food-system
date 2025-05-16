CREATE TABLE public.station (
    station_id uuid DEFAULT uuid_generate_v4() NOT NULL,
    name varchar(255) NOT NULL,
    created_at timestamptz DEFAULT now()
);