SET search_path TO app, public;

ALTER TABLE "order"
  ADD COLUMN payment_attempt_id VARCHAR(100),
  ADD COLUMN payrexx_gateway_id INTEGER,
  ADD COLUMN payrexx_transaction_id INTEGER;

CREATE UNIQUE INDEX idx_order_payment_attempt_pending
  ON "order" (payment_attempt_id)
  WHERE status = 'pending' AND payment_attempt_id IS NOT NULL;

CREATE INDEX idx_order_payrexx_gateway ON "order" (payrexx_gateway_id) WHERE payrexx_gateway_id IS NOT NULL;
CREATE INDEX idx_order_payrexx_transaction ON "order" (payrexx_transaction_id) WHERE payrexx_transaction_id IS NOT NULL;
