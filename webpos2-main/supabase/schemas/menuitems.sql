CREATE TABLE public.menuitems (
    menu_id uuid NOT NULL,
    product_id uuid NOT NULL,
    quantity integer DEFAULT 1 NOT NULL,
    status item_status DEFAULT 'Available'::item_status NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    sequence integer DEFAULT 1
);