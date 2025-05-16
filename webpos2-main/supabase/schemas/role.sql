CREATE TABLE public.role (
    role_id uuid DEFAULT uuid_generate_v4() NOT NULL,
    role_name varchar(50) NOT NULL,
    description text,
    created_at timestamptz DEFAULT now()
);