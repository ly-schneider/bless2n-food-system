CREATE TABLE public.order_status (
    status_id uuid DEFAULT uuid_generate_v4() NOT NULL,
    status_name varchar(50) NOT NULL,
    created_at timestamptz DEFAULT now()
);