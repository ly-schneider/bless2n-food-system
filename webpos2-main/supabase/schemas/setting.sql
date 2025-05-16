CREATE TABLE public.setting (
    setting_id uuid DEFAULT uuid_generate_v4() NOT NULL,
    setting_name text NOT NULL,
    value text NOT NULL,
    created_at timestamptz DEFAULT now()
);