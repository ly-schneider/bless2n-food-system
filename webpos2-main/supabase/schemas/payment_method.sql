CREATE TABLE public.payment_method (
    payment_method_id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    description text,
    payment_code text NOT NULL
);