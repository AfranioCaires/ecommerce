-- name: CreateCustomer :exec
INSERT INTO customers (id, name, email, password_hash, roles, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetCustomerByEmail :one
SELECT id, name, email, password_hash, roles, created_at
FROM customers
WHERE email = $1;

-- name: GetCustomerByID :one
SELECT id, name, email, password_hash, roles, created_at
FROM customers
WHERE id = $1;

-- name: ListCustomers :many
SELECT id, name, email, password_hash, roles, created_at
FROM customers
ORDER BY created_at, id;

