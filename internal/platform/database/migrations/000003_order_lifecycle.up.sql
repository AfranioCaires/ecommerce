ALTER TABLE orders DROP CONSTRAINT orders_status_check;
UPDATE orders SET status = 'CANCELED' WHERE status = 'FAILED';
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('PENDING', 'PAYMENT_PENDING', 'PAID', 'CANCELED'));
