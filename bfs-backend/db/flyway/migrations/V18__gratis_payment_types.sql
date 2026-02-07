SET search_path TO app, public;

ALTER TYPE payment_method ADD VALUE 'GRATIS_GUEST';
ALTER TYPE payment_method ADD VALUE 'GRATIS_VIP';
ALTER TYPE payment_method ADD VALUE 'GRATIS_100CLUB';
