-- name: CreateCustomer :exec
INSERT INTO customers (id, email, password_hash, roles, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetCustomerByEmail :one
SELECT id, email, password_hash, roles, created_at
FROM customers
WHERE email = $1;

-- name: GetCustomerByID :one
SELECT id, email, password_hash, roles, created_at
FROM customers
WHERE id = $1;

