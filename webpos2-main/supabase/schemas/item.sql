CREATE TABLE public.item (
    item_id uuid DEFAULT uuid_generate_v4() NOT NULL,
    name varchar(255) NOT NULL,
    type item_type NOT NULL,
    status item_status NOT NULL,
    stock integer,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    emoji text,
    categroy_id uuid,
    price numeric,
    sequence integer DEFAULT 1,
    color_hex text,
    location_id uuid
);