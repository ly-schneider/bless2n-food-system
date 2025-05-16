CREATE TABLE public.company (
    company_id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    name text NOT NULL
);