-- name: UpsertStock :exec
INSERT INTO stocks (product_id, quantity, updated_at)
VALUES ($1, $2, $3)
ON CONFLICT (product_id) DO UPDATE SET
    quantity = EXCLUDED.quantity,
    updated_at = EXCLUDED.updated_at;

-- name: GetStockByProductID :one
SELECT product_id, quantity, updated_at
FROM stocks
WHERE product_id = $1;

-- name: GetStockByProductIDForUpdate :one
SELECT product_id, quantity, updated_at
FROM stocks
WHERE product_id = $1
FOR UPDATE;

