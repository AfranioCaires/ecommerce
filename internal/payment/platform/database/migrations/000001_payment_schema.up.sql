CREATE TABLE payments (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL UNIQUE,
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    status TEXT NOT NULL CHECK (status IN ('APPROVED', 'DECLINED')),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE outbox_messages (
    id TEXT PRIMARY KEY,
    message_type TEXT NOT NULL,
    routing_key TEXT NOT NULL,
    payload JSONB NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL,
    published_at TIMESTAMPTZ,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX payment_outbox_pending_idx ON outbox_messages (next_attempt_at, created_at) WHERE published_at IS NULL;

CREATE TABLE inbox_messages (
    message_id TEXT PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL
);
