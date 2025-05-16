CREATE TABLE public.item_category (
    item_category_id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    description text,
    location_id uuid
);