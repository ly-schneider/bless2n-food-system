CREATE TABLE public.order (
    order_id uuid DEFAULT uuid_generate_v4() NOT NULL,
    order_number integer NOT NULL,
    user_id uuid,
    status_id uuid DEFAULT 'f70b7812-b50b-4079-ac3b-e238d1abfb71'::uuid,
    order_date timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    pickup_date timestamptz,
    donation_amount numeric(10,2) DEFAULT 0,
    total_amount numeric(10,2) DEFAULT 0 NOT NULL,
    created_at timestamptz DEFAULT now(),
    discount_code text,
    discount_amount numeric,
    payment_method_id uuid NOT NULL
);