-- name: CreatePaymentOutboxMessage :exec
INSERT INTO outbox_messages (id, message_type, routing_key, payload, attempts, next_attempt_at, published_at, last_error, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: ListPendingPaymentOutboxMessages :many
SELECT id, message_type, routing_key, payload, attempts, next_attempt_at, published_at, last_error, created_at
FROM outbox_messages WHERE published_at IS NULL AND next_attempt_at <= $1
ORDER BY created_at LIMIT $2 FOR UPDATE SKIP LOCKED;

-- name: MarkPaymentOutboxPublished :exec
UPDATE outbox_messages SET published_at = $2, last_error = '' WHERE id = $1;

-- name: MarkPaymentOutboxFailed :exec
UPDATE outbox_messages SET attempts = attempts + 1, next_attempt_at = $2, last_error = $3 WHERE id = $1;

-- name: CreatePaymentInboxMessage :exec
INSERT INTO inbox_messages (message_id, processed_at) VALUES ($1, $2);
