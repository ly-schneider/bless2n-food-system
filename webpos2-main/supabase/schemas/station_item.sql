CREATE TABLE public.station_item (
    station_id uuid NOT NULL,
    item_id uuid NOT NULL,
    created_at timestamptz DEFAULT now()
);