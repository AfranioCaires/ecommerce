ALTER TABLE orders DROP CONSTRAINT orders_status_check;
UPDATE orders SET status = 'PENDING' WHERE status = 'PAYMENT_PENDING';
UPDATE orders SET status = 'FAILED' WHERE status = 'CANCELED';
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('PENDING', 'PAID', 'FAILED'));
