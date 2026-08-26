CREATE TABLE payments (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    status TEXT NOT NULL CHECK (status IN ('APPROVED', 'DECLINED')),
    created_at TIMESTAMPTZ NOT NULL
);
