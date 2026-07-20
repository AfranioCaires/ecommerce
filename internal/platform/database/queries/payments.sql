-- name: CreatePayment :exec
INSERT INTO payments (
    id, order_id, amount_cents, status, created_at
)
VALUES ($1, $2, $3, $4, $5);

