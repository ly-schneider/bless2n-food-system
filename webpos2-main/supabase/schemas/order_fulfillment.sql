CREATE TABLE public.order_fulfillment (
    order_fulfillment_id uuid DEFAULT uuid_generate_v4() NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    order_id uuid NOT NULL,
    item_id uuid NOT NULL,
    order_line_id uuid,
    status_id uuid NOT NULL
);