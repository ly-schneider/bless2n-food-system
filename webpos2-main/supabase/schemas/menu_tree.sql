CREATE TABLE public.menu_tree (
    menu_tree_id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    text text,
    role_id uuid
);