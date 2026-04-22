ALTER TABLE order_payment
    ADD COLUMN card_brand          VARCHAR NULL,
    ADD COLUMN card_last4          VARCHAR NULL,
    ADD COLUMN entry_mode          VARCHAR NULL,
    ADD COLUMN card_transaction_id VARCHAR NULL;
