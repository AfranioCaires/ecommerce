-- name: CreateOrder :exec
INSERT INTO orders (
    id, customer_id, total_amount_cents, status, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: CreateOrderItem :exec
INSERT INTO order_items (
    order_id, product_id, product_name, unit_price_cents, quantity
)
VALUES ($1, $2, $3, $4, $5);

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2, updated_at = $3
WHERE id = $1;

-- name: GetOrderByID :one
SELECT id, customer_id, total_amount_cents, status, created_at, updated_at
FROM orders
WHERE id = $1;

-- name: GetOrderByIDForUpdate :one
SELECT id, customer_id, total_amount_cents, status, created_at, updated_at
FROM orders
WHERE id = $1
FOR UPDATE;

-- name: ListOrdersByCustomer :many
SELECT id, customer_id, total_amount_cents, status, created_at, updated_at
FROM orders
WHERE customer_id = $1
ORDER BY created_at DESC
LIMIT sqlc.arg(page_limit)::integer
OFFSET sqlc.arg(page_offset)::integer;

-- name: ListOrders :many
SELECT id, customer_id, total_amount_cents, status, created_at, updated_at
FROM orders
ORDER BY created_at DESC
LIMIT sqlc.arg(page_limit)::integer
OFFSET sqlc.arg(page_offset)::integer;

-- name: ListOrderItemsByOrderIDs :many
SELECT id, order_id, product_id, product_name, unit_price_cents, quantity
FROM order_items
WHERE order_id = ANY(sqlc.arg(order_ids)::text[])
ORDER BY id;

