create type "public"."order_status" as enum ('pending', 'completed', 'cancelled');

create table "public"."category" (
    "id" uuid not null default uuid_generate_v4(),
    "event_id" uuid not null,
    "name" text not null,
    "emoji" text not null,
    "is_active" boolean not null default true,
    "created_at" timestamp with time zone not null default now()
);


alter table "public"."category" enable row level security;

create table "public"."devices" (
    "id" uuid not null default uuid_generate_v4(),
    "serial_number" text not null,
    "model" text not null,
    "is_active" boolean not null default true,
    "registered_at" timestamp with time zone not null default now()
);


alter table "public"."devices" enable row level security;

create table "public"."event_device_pairings" (
    "id" uuid not null default uuid_generate_v4(),
    "customer_device_id" uuid,
    "cashier_device_id" uuid,
    "linked_at" timestamp with time zone not null
);


alter table "public"."event_device_pairings" enable row level security;

create table "public"."event_devices" (
    "id" uuid not null default uuid_generate_v4(),
    "event_id" uuid not null,
    "device_id" uuid not null,
    "assigned_at" timestamp with time zone not null default now()
);


alter table "public"."event_devices" enable row level security;

create table "public"."event_invites" (
    "id" uuid not null default uuid_generate_v4(),
    "event_id" uuid not null,
    "role_id" uuid not null,
    "invitee_email" text not null,
    "expires_at" timestamp with time zone not null
);


alter table "public"."event_invites" enable row level security;

create table "public"."event_user_roles" (
    "id" uuid not null default uuid_generate_v4(),
    "name" text not null,
    "display_name" text not null,
    "description" text,
    "is_default" boolean not null default false,
    "is_active" boolean not null default true,
    "created_at" timestamp with time zone not null default now()
);


alter table "public"."event_user_roles" enable row level security;

create table "public"."event_users" (
    "id" uuid not null default uuid_generate_v4(),
    "event_id" uuid not null,
    "user_id" uuid not null,
    "role_id" uuid not null,
    "invited_at" timestamp with time zone not null default now(),
    "joined_at" timestamp with time zone not null
);


alter table "public"."event_users" enable row level security;

create table "public"."events" (
    "id" uuid not null default uuid_generate_v4(),
    "owner_id" uuid not null,
    "name" text not null,
    "start_date" date not null,
    "end_date" date not null,
    "created_at" timestamp with time zone not null default now()
);


alter table "public"."events" enable row level security;

create table "public"."order_items" (
    "id" uuid not null default uuid_generate_v4(),
    "order_id" uuid not null,
    "product_id" uuid not null,
    "quantity" integer not null,
    "price_per_unit" numeric(12,2) not null
);


alter table "public"."order_items" enable row level security;

create table "public"."orders" (
    "id" uuid not null default uuid_generate_v4(),
    "event_id" uuid not null,
    "device_id" uuid not null,
    "total" numeric(12,2) not null,
    "status" order_status not null,
    "ordered_at" timestamp with time zone not null default now()
);


alter table "public"."orders" enable row level security;

create table "public"."products" (
    "id" uuid not null default uuid_generate_v4(),
    "event_id" uuid not null,
    "category_id" uuid,
    "name" text not null,
    "emoji" text not null,
    "price" numeric(12,2),
    "is_active" boolean not null default true,
    "created_at" timestamp with time zone not null default now()
);


alter table "public"."products" enable row level security;

create table "public"."users" (
    "id" uuid not null,
    "firstname" text not null,
    "lastname" text not null,
    "email" text not null,
    "is_site_admin" boolean not null default false,
    "is_disabled" boolean not null default false,
    "disabled_reason" text,
    "created_at" timestamp with time zone not null default now()
);


alter table "public"."users" enable row level security;

CREATE UNIQUE INDEX category_event_name_ci_key ON public.category USING btree (event_id, lower(name));

CREATE UNIQUE INDEX category_pkey ON public.category USING btree (id);

CREATE UNIQUE INDEX devices_pkey ON public.devices USING btree (id);

CREATE UNIQUE INDEX devices_serial_number_key ON public.devices USING btree (serial_number);

CREATE UNIQUE INDEX event_device_pairings_cashier_device_id_key ON public.event_device_pairings USING btree (cashier_device_id);

CREATE UNIQUE INDEX event_device_pairings_customer_device_id_key ON public.event_device_pairings USING btree (customer_device_id);

CREATE UNIQUE INDEX event_device_pairings_pkey ON public.event_device_pairings USING btree (id);

CREATE UNIQUE INDEX event_devices_event_id_device_id_key ON public.event_devices USING btree (event_id, device_id);

CREATE UNIQUE INDEX event_devices_pkey ON public.event_devices USING btree (id);

CREATE UNIQUE INDEX event_invites_pkey ON public.event_invites USING btree (id);

CREATE UNIQUE INDEX event_user_roles_name_key ON public.event_user_roles USING btree (name);

CREATE UNIQUE INDEX event_user_roles_pkey ON public.event_user_roles USING btree (id);

CREATE UNIQUE INDEX event_users_event_id_user_id_key ON public.event_users USING btree (event_id, user_id);

CREATE UNIQUE INDEX event_users_pkey ON public.event_users USING btree (id);

CREATE UNIQUE INDEX events_pkey ON public.events USING btree (id);

CREATE UNIQUE INDEX order_items_order_id_product_id_key ON public.order_items USING btree (order_id, product_id);

CREATE UNIQUE INDEX order_items_pkey ON public.order_items USING btree (id);

CREATE UNIQUE INDEX orders_pkey ON public.orders USING btree (id);

CREATE UNIQUE INDEX products_event_name_ci_key ON public.products USING btree (event_id, lower(name));

CREATE UNIQUE INDEX products_pkey ON public.products USING btree (id);

CREATE UNIQUE INDEX users_email_key ON public.users USING btree (email);

CREATE UNIQUE INDEX users_pkey ON public.users USING btree (id);

alter table "public"."category" add constraint "category_pkey" PRIMARY KEY using index "category_pkey";

alter table "public"."devices" add constraint "devices_pkey" PRIMARY KEY using index "devices_pkey";

alter table "public"."event_device_pairings" add constraint "event_device_pairings_pkey" PRIMARY KEY using index "event_device_pairings_pkey";

alter table "public"."event_devices" add constraint "event_devices_pkey" PRIMARY KEY using index "event_devices_pkey";

alter table "public"."event_invites" add constraint "event_invites_pkey" PRIMARY KEY using index "event_invites_pkey";

alter table "public"."event_user_roles" add constraint "event_user_roles_pkey" PRIMARY KEY using index "event_user_roles_pkey";

alter table "public"."event_users" add constraint "event_users_pkey" PRIMARY KEY using index "event_users_pkey";

alter table "public"."events" add constraint "events_pkey" PRIMARY KEY using index "events_pkey";

alter table "public"."order_items" add constraint "order_items_pkey" PRIMARY KEY using index "order_items_pkey";

alter table "public"."orders" add constraint "orders_pkey" PRIMARY KEY using index "orders_pkey";

alter table "public"."products" add constraint "products_pkey" PRIMARY KEY using index "products_pkey";

alter table "public"."users" add constraint "users_pkey" PRIMARY KEY using index "users_pkey";

alter table "public"."category" add constraint "category_event_id_fkey" FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE not valid;

alter table "public"."category" validate constraint "category_event_id_fkey";

alter table "public"."devices" add constraint "devices_serial_number_key" UNIQUE using index "devices_serial_number_key";

alter table "public"."event_device_pairings" add constraint "event_device_pairings_cashier_device_id_fkey" FOREIGN KEY (cashier_device_id) REFERENCES event_devices(id) ON DELETE CASCADE not valid;

alter table "public"."event_device_pairings" validate constraint "event_device_pairings_cashier_device_id_fkey";

alter table "public"."event_device_pairings" add constraint "event_device_pairings_cashier_device_id_key" UNIQUE using index "event_device_pairings_cashier_device_id_key";

alter table "public"."event_device_pairings" add constraint "event_device_pairings_customer_device_id_fkey" FOREIGN KEY (customer_device_id) REFERENCES event_devices(id) ON DELETE CASCADE not valid;

alter table "public"."event_device_pairings" validate constraint "event_device_pairings_customer_device_id_fkey";

alter table "public"."event_device_pairings" add constraint "event_device_pairings_customer_device_id_key" UNIQUE using index "event_device_pairings_customer_device_id_key";

alter table "public"."event_devices" add constraint "event_devices_device_id_fkey" FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE not valid;

alter table "public"."event_devices" validate constraint "event_devices_device_id_fkey";

alter table "public"."event_devices" add constraint "event_devices_event_id_device_id_key" UNIQUE using index "event_devices_event_id_device_id_key";

alter table "public"."event_devices" add constraint "event_devices_event_id_fkey" FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE not valid;

alter table "public"."event_devices" validate constraint "event_devices_event_id_fkey";

alter table "public"."event_invites" add constraint "event_invites_event_id_fkey" FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE not valid;

alter table "public"."event_invites" validate constraint "event_invites_event_id_fkey";

alter table "public"."event_invites" add constraint "event_invites_role_id_fkey" FOREIGN KEY (role_id) REFERENCES event_user_roles(id) not valid;

alter table "public"."event_invites" validate constraint "event_invites_role_id_fkey";

alter table "public"."event_user_roles" add constraint "event_user_roles_name_check" CHECK ((name = lower(name))) not valid;

alter table "public"."event_user_roles" validate constraint "event_user_roles_name_check";

alter table "public"."event_user_roles" add constraint "event_user_roles_name_key" UNIQUE using index "event_user_roles_name_key";

alter table "public"."event_users" add constraint "event_users_event_id_fkey" FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE not valid;

alter table "public"."event_users" validate constraint "event_users_event_id_fkey";

alter table "public"."event_users" add constraint "event_users_event_id_user_id_key" UNIQUE using index "event_users_event_id_user_id_key";

alter table "public"."event_users" add constraint "event_users_role_id_fkey" FOREIGN KEY (role_id) REFERENCES event_user_roles(id) not valid;

alter table "public"."event_users" validate constraint "event_users_role_id_fkey";

alter table "public"."event_users" add constraint "event_users_user_id_fkey" FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE not valid;

alter table "public"."event_users" validate constraint "event_users_user_id_fkey";

alter table "public"."events" add constraint "events_check" CHECK ((end_date >= start_date)) not valid;

alter table "public"."events" validate constraint "events_check";

alter table "public"."events" add constraint "events_owner_id_fkey" FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE not valid;

alter table "public"."events" validate constraint "events_owner_id_fkey";

alter table "public"."order_items" add constraint "order_items_order_id_fkey" FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE not valid;

alter table "public"."order_items" validate constraint "order_items_order_id_fkey";

alter table "public"."order_items" add constraint "order_items_order_id_product_id_key" UNIQUE using index "order_items_order_id_product_id_key";

alter table "public"."order_items" add constraint "order_items_price_per_unit_check" CHECK ((price_per_unit >= (0)::numeric)) not valid;

alter table "public"."order_items" validate constraint "order_items_price_per_unit_check";

alter table "public"."order_items" add constraint "order_items_product_id_fkey" FOREIGN KEY (product_id) REFERENCES products(id) not valid;

alter table "public"."order_items" validate constraint "order_items_product_id_fkey";

alter table "public"."order_items" add constraint "order_items_quantity_check" CHECK ((quantity > 0)) not valid;

alter table "public"."order_items" validate constraint "order_items_quantity_check";

alter table "public"."orders" add constraint "orders_device_id_fkey" FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE SET NULL not valid;

alter table "public"."orders" validate constraint "orders_device_id_fkey";

alter table "public"."orders" add constraint "orders_event_id_fkey" FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE not valid;

alter table "public"."orders" validate constraint "orders_event_id_fkey";

alter table "public"."orders" add constraint "orders_total_check" CHECK ((total >= (0)::numeric)) not valid;

alter table "public"."orders" validate constraint "orders_total_check";

alter table "public"."products" add constraint "products_category_id_fkey" FOREIGN KEY (category_id) REFERENCES category(id) not valid;

alter table "public"."products" validate constraint "products_category_id_fkey";

alter table "public"."products" add constraint "products_event_id_fkey" FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE not valid;

alter table "public"."products" validate constraint "products_event_id_fkey";

alter table "public"."products" add constraint "products_price_check" CHECK ((price >= (0)::numeric)) not valid;

alter table "public"."products" validate constraint "products_price_check";

alter table "public"."users" add constraint "users_email_key" UNIQUE using index "users_email_key";

alter table "public"."users" add constraint "users_id_fkey" FOREIGN KEY (id) REFERENCES auth.users(id) ON DELETE CASCADE not valid;

alter table "public"."users" validate constraint "users_id_fkey";

grant delete on table "public"."category" to "anon";

grant insert on table "public"."category" to "anon";

grant references on table "public"."category" to "anon";

grant select on table "public"."category" to "anon";

grant trigger on table "public"."category" to "anon";

grant truncate on table "public"."category" to "anon";

grant update on table "public"."category" to "anon";

grant delete on table "public"."category" to "authenticated";

grant insert on table "public"."category" to "authenticated";

grant references on table "public"."category" to "authenticated";

grant select on table "public"."category" to "authenticated";

grant trigger on table "public"."category" to "authenticated";

grant truncate on table "public"."category" to "authenticated";

grant update on table "public"."category" to "authenticated";

grant delete on table "public"."category" to "service_role";

grant insert on table "public"."category" to "service_role";

grant references on table "public"."category" to "service_role";

grant select on table "public"."category" to "service_role";

grant trigger on table "public"."category" to "service_role";

grant truncate on table "public"."category" to "service_role";

grant update on table "public"."category" to "service_role";

grant delete on table "public"."devices" to "anon";

grant insert on table "public"."devices" to "anon";

grant references on table "public"."devices" to "anon";

grant select on table "public"."devices" to "anon";

grant trigger on table "public"."devices" to "anon";

grant truncate on table "public"."devices" to "anon";

grant update on table "public"."devices" to "anon";

grant delete on table "public"."devices" to "authenticated";

grant insert on table "public"."devices" to "authenticated";

grant references on table "public"."devices" to "authenticated";

grant select on table "public"."devices" to "authenticated";

grant trigger on table "public"."devices" to "authenticated";

grant truncate on table "public"."devices" to "authenticated";

grant update on table "public"."devices" to "authenticated";

grant delete on table "public"."devices" to "service_role";

grant insert on table "public"."devices" to "service_role";

grant references on table "public"."devices" to "service_role";

grant select on table "public"."devices" to "service_role";

grant trigger on table "public"."devices" to "service_role";

grant truncate on table "public"."devices" to "service_role";

grant update on table "public"."devices" to "service_role";

grant delete on table "public"."event_device_pairings" to "anon";

grant insert on table "public"."event_device_pairings" to "anon";

grant references on table "public"."event_device_pairings" to "anon";

grant select on table "public"."event_device_pairings" to "anon";

grant trigger on table "public"."event_device_pairings" to "anon";

grant truncate on table "public"."event_device_pairings" to "anon";

grant update on table "public"."event_device_pairings" to "anon";

grant delete on table "public"."event_device_pairings" to "authenticated";

grant insert on table "public"."event_device_pairings" to "authenticated";

grant references on table "public"."event_device_pairings" to "authenticated";

grant select on table "public"."event_device_pairings" to "authenticated";

grant trigger on table "public"."event_device_pairings" to "authenticated";

grant truncate on table "public"."event_device_pairings" to "authenticated";

grant update on table "public"."event_device_pairings" to "authenticated";

grant delete on table "public"."event_device_pairings" to "service_role";

grant insert on table "public"."event_device_pairings" to "service_role";

grant references on table "public"."event_device_pairings" to "service_role";

grant select on table "public"."event_device_pairings" to "service_role";

grant trigger on table "public"."event_device_pairings" to "service_role";

grant truncate on table "public"."event_device_pairings" to "service_role";

grant update on table "public"."event_device_pairings" to "service_role";

grant delete on table "public"."event_devices" to "anon";

grant insert on table "public"."event_devices" to "anon";

grant references on table "public"."event_devices" to "anon";

grant select on table "public"."event_devices" to "anon";

grant trigger on table "public"."event_devices" to "anon";

grant truncate on table "public"."event_devices" to "anon";

grant update on table "public"."event_devices" to "anon";

grant delete on table "public"."event_devices" to "authenticated";

grant insert on table "public"."event_devices" to "authenticated";

grant references on table "public"."event_devices" to "authenticated";

grant select on table "public"."event_devices" to "authenticated";

grant trigger on table "public"."event_devices" to "authenticated";

grant truncate on table "public"."event_devices" to "authenticated";

grant update on table "public"."event_devices" to "authenticated";

grant delete on table "public"."event_devices" to "service_role";

grant insert on table "public"."event_devices" to "service_role";

grant references on table "public"."event_devices" to "service_role";

grant select on table "public"."event_devices" to "service_role";

grant trigger on table "public"."event_devices" to "service_role";

grant truncate on table "public"."event_devices" to "service_role";

grant update on table "public"."event_devices" to "service_role";

grant delete on table "public"."event_invites" to "anon";

grant insert on table "public"."event_invites" to "anon";

grant references on table "public"."event_invites" to "anon";

grant select on table "public"."event_invites" to "anon";

grant trigger on table "public"."event_invites" to "anon";

grant truncate on table "public"."event_invites" to "anon";

grant update on table "public"."event_invites" to "anon";

grant delete on table "public"."event_invites" to "authenticated";

grant insert on table "public"."event_invites" to "authenticated";

grant references on table "public"."event_invites" to "authenticated";

grant select on table "public"."event_invites" to "authenticated";

grant trigger on table "public"."event_invites" to "authenticated";

grant truncate on table "public"."event_invites" to "authenticated";

grant update on table "public"."event_invites" to "authenticated";

grant delete on table "public"."event_invites" to "service_role";

grant insert on table "public"."event_invites" to "service_role";

grant references on table "public"."event_invites" to "service_role";

grant select on table "public"."event_invites" to "service_role";

grant trigger on table "public"."event_invites" to "service_role";

grant truncate on table "public"."event_invites" to "service_role";

grant update on table "public"."event_invites" to "service_role";

grant delete on table "public"."event_user_roles" to "anon";

grant insert on table "public"."event_user_roles" to "anon";

grant references on table "public"."event_user_roles" to "anon";

grant select on table "public"."event_user_roles" to "anon";

grant trigger on table "public"."event_user_roles" to "anon";

grant truncate on table "public"."event_user_roles" to "anon";

grant update on table "public"."event_user_roles" to "anon";

grant delete on table "public"."event_user_roles" to "authenticated";

grant insert on table "public"."event_user_roles" to "authenticated";

grant references on table "public"."event_user_roles" to "authenticated";

grant select on table "public"."event_user_roles" to "authenticated";

grant trigger on table "public"."event_user_roles" to "authenticated";

grant truncate on table "public"."event_user_roles" to "authenticated";

grant update on table "public"."event_user_roles" to "authenticated";

grant delete on table "public"."event_user_roles" to "service_role";

grant insert on table "public"."event_user_roles" to "service_role";

grant references on table "public"."event_user_roles" to "service_role";

grant select on table "public"."event_user_roles" to "service_role";

grant trigger on table "public"."event_user_roles" to "service_role";

grant truncate on table "public"."event_user_roles" to "service_role";

grant update on table "public"."event_user_roles" to "service_role";

grant delete on table "public"."event_users" to "anon";

grant insert on table "public"."event_users" to "anon";

grant references on table "public"."event_users" to "anon";

grant select on table "public"."event_users" to "anon";

grant trigger on table "public"."event_users" to "anon";

grant truncate on table "public"."event_users" to "anon";

grant update on table "public"."event_users" to "anon";

grant delete on table "public"."event_users" to "authenticated";

grant insert on table "public"."event_users" to "authenticated";

grant references on table "public"."event_users" to "authenticated";

grant select on table "public"."event_users" to "authenticated";

grant trigger on table "public"."event_users" to "authenticated";

grant truncate on table "public"."event_users" to "authenticated";

grant update on table "public"."event_users" to "authenticated";

grant delete on table "public"."event_users" to "service_role";

grant insert on table "public"."event_users" to "service_role";

grant references on table "public"."event_users" to "service_role";

grant select on table "public"."event_users" to "service_role";

grant trigger on table "public"."event_users" to "service_role";

grant truncate on table "public"."event_users" to "service_role";

grant update on table "public"."event_users" to "service_role";

grant delete on table "public"."events" to "anon";

grant insert on table "public"."events" to "anon";

grant references on table "public"."events" to "anon";

grant select on table "public"."events" to "anon";

grant trigger on table "public"."events" to "anon";

grant truncate on table "public"."events" to "anon";

grant update on table "public"."events" to "anon";

grant delete on table "public"."events" to "authenticated";

grant insert on table "public"."events" to "authenticated";

grant references on table "public"."events" to "authenticated";

grant select on table "public"."events" to "authenticated";

grant trigger on table "public"."events" to "authenticated";

grant truncate on table "public"."events" to "authenticated";

grant update on table "public"."events" to "authenticated";

grant delete on table "public"."events" to "service_role";

grant insert on table "public"."events" to "service_role";

grant references on table "public"."events" to "service_role";

grant select on table "public"."events" to "service_role";

grant trigger on table "public"."events" to "service_role";

grant truncate on table "public"."events" to "service_role";

grant update on table "public"."events" to "service_role";

grant delete on table "public"."order_items" to "anon";

grant insert on table "public"."order_items" to "anon";

grant references on table "public"."order_items" to "anon";

grant select on table "public"."order_items" to "anon";

grant trigger on table "public"."order_items" to "anon";

grant truncate on table "public"."order_items" to "anon";

grant update on table "public"."order_items" to "anon";

grant delete on table "public"."order_items" to "authenticated";

grant insert on table "public"."order_items" to "authenticated";

grant references on table "public"."order_items" to "authenticated";

grant select on table "public"."order_items" to "authenticated";

grant trigger on table "public"."order_items" to "authenticated";

grant truncate on table "public"."order_items" to "authenticated";

grant update on table "public"."order_items" to "authenticated";

grant delete on table "public"."order_items" to "service_role";

grant insert on table "public"."order_items" to "service_role";

grant references on table "public"."order_items" to "service_role";

grant select on table "public"."order_items" to "service_role";

grant trigger on table "public"."order_items" to "service_role";

grant truncate on table "public"."order_items" to "service_role";

grant update on table "public"."order_items" to "service_role";

grant delete on table "public"."orders" to "anon";

grant insert on table "public"."orders" to "anon";

grant references on table "public"."orders" to "anon";

grant select on table "public"."orders" to "anon";

grant trigger on table "public"."orders" to "anon";

grant truncate on table "public"."orders" to "anon";

grant update on table "public"."orders" to "anon";

grant delete on table "public"."orders" to "authenticated";

grant insert on table "public"."orders" to "authenticated";

grant references on table "public"."orders" to "authenticated";

grant select on table "public"."orders" to "authenticated";

grant trigger on table "public"."orders" to "authenticated";

grant truncate on table "public"."orders" to "authenticated";

grant update on table "public"."orders" to "authenticated";

grant delete on table "public"."orders" to "service_role";

grant insert on table "public"."orders" to "service_role";

grant references on table "public"."orders" to "service_role";

grant select on table "public"."orders" to "service_role";

grant trigger on table "public"."orders" to "service_role";

grant truncate on table "public"."orders" to "service_role";

grant update on table "public"."orders" to "service_role";

grant delete on table "public"."products" to "anon";

grant insert on table "public"."products" to "anon";

grant references on table "public"."products" to "anon";

grant select on table "public"."products" to "anon";

grant trigger on table "public"."products" to "anon";

grant truncate on table "public"."products" to "anon";

grant update on table "public"."products" to "anon";

grant delete on table "public"."products" to "authenticated";

grant insert on table "public"."products" to "authenticated";

grant references on table "public"."products" to "authenticated";

grant select on table "public"."products" to "authenticated";

grant trigger on table "public"."products" to "authenticated";

grant truncate on table "public"."products" to "authenticated";

grant update on table "public"."products" to "authenticated";

grant delete on table "public"."products" to "service_role";

grant insert on table "public"."products" to "service_role";

grant references on table "public"."products" to "service_role";

grant select on table "public"."products" to "service_role";

grant trigger on table "public"."products" to "service_role";

grant truncate on table "public"."products" to "service_role";

grant update on table "public"."products" to "service_role";

grant delete on table "public"."users" to "anon";

grant insert on table "public"."users" to "anon";

grant references on table "public"."users" to "anon";

grant select on table "public"."users" to "anon";

grant trigger on table "public"."users" to "anon";

grant truncate on table "public"."users" to "anon";

grant update on table "public"."users" to "anon";

grant delete on table "public"."users" to "authenticated";

grant insert on table "public"."users" to "authenticated";

grant references on table "public"."users" to "authenticated";

grant select on table "public"."users" to "authenticated";

grant trigger on table "public"."users" to "authenticated";

grant truncate on table "public"."users" to "authenticated";

grant update on table "public"."users" to "authenticated";

grant delete on table "public"."users" to "service_role";

grant insert on table "public"."users" to "service_role";

grant references on table "public"."users" to "service_role";

grant select on table "public"."users" to "service_role";

grant trigger on table "public"."users" to "service_role";

grant truncate on table "public"."users" to "service_role";

grant update on table "public"."users" to "service_role";


