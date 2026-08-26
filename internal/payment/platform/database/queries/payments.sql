-- name: CreatePayment :exec
INSERT INTO payments (id, order_id, amount_cents, status, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetPaymentByOrderID :one
SELECT id, order_id, amount_cents, status, created_at
FROM payments WHERE order_id = $1;
