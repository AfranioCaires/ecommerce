-- name: UpsertProduct :exec
INSERT INTO products (
    id, name, description, price_cents, status, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price_cents = EXCLUDED.price_cents,
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at;

-- name: GetProductByID :one
SELECT id, name, description, price_cents, status, created_at, updated_at
FROM products
WHERE id = $1;

-- name: ListProductsByIDs :many
SELECT id, name, description, price_cents, status, created_at, updated_at
FROM products
WHERE id = ANY(sqlc.arg(product_ids)::text[]);

-- name: CountActiveProducts :one
SELECT COUNT(*)
FROM products
WHERE status = 'ACTIVE';

-- name: ListActiveProducts :many
SELECT id, name, description, price_cents, status, created_at, updated_at
FROM products
WHERE status = 'ACTIVE'
ORDER BY created_at DESC
LIMIT sqlc.arg(page_limit)::integer
OFFSET sqlc.arg(page_offset)::integer;

