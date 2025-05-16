CREATE TABLE public.location (
    location_id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    description text NOT NULL,
    name text,
    street text,
    city text,
    postcode text,
    company_id uuid
);