SET search_path TO app, public;

ALTER TYPE inventory_reason ADD VALUE IF NOT EXISTS 'cancellation' AFTER 'refund';
