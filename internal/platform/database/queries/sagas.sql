-- name: CreateOrderSaga :exec
INSERT INTO order_sagas (id, order_id, correlation_id, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetOrderSagaForUpdate :one
SELECT id, order_id, correlation_id, status, created_at, updated_at
FROM order_sagas
WHERE id = $1
FOR UPDATE;

-- name: UpdateOrderSaga :exec
UPDATE order_sagas SET status = $2, updated_at = $3 WHERE id = $1;
