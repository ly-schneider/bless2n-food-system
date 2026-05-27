DELETE FROM order_line_redemption a
USING order_line_redemption b
WHERE a.order_line_id = b.order_line_id
  AND (a.redeemed_at, a.id) > (b.redeemed_at, b.id);

DROP INDEX IF EXISTS idx_order_line_redemption_order_line_id;

CREATE UNIQUE INDEX idx_order_line_redemption_order_line_id ON order_line_redemption (order_line_id);
