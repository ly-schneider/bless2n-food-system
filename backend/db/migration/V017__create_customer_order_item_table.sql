CREATE TABLE customer_order_item (
    customer_order_id nano_id NOT NULL,
    event_product_id nano_id NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    price_per_unit NUMERIC(6, 2) NOT NULL CHECK (price_per_unit >= 0),
    FOREIGN KEY (customer_order_id) REFERENCES customer_order (id) ON DELETE CASCADE,
    FOREIGN KEY (event_product_id) REFERENCES event_product (id) ON DELETE CASCADE
);