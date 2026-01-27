CREATE TYPE user_role AS ENUM ('admin', 'customer');

CREATE TYPE order_status AS ENUM ('pending', 'paid', 'cancelled', 'refunded');

CREATE TYPE order_origin AS ENUM ('shop', 'pos');

CREATE TYPE product_type AS ENUM ('simple', 'menu');

CREATE TYPE order_item_type AS ENUM ('simple', 'bundle', 'component');

CREATE TYPE inventory_reason AS ENUM (
  'opening_balance',
  'sale',
  'refund',
  'manual_adjust',
  'correction'
);

CREATE TYPE common_status AS ENUM ('pending', 'approved', 'rejected', 'revoked');

CREATE TYPE pos_fulfillment_mode AS ENUM ('QR_CODE', 'JETON');

CREATE TYPE payment_method AS ENUM ('CASH', 'CARD', 'TWINT');

CREATE TYPE device_type AS ENUM ('POS', 'STATION');