CREATE TABLE public.order_line (
    order_id uuid NOT NULL,
    item_id uuid NOT NULL,
    quantity integer NOT NULL,
    price numeric NOT NULL,
    order_line_id uuid DEFAULT uuid_generate_v4() NOT NULL,
    line_amount numeric,
    created_at timestamptz DEFAULT now(),
    description text
);