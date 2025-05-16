CREATE TABLE public.profile (
    user_id uuid NOT NULL,
    email text NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    name text,
    location_id uuid
);