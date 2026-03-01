CREATE TABLE IF NOT EXISTS webhook_deliveries (
    delivery_id      TEXT PRIMARY KEY NOT NULL CHECK (length(delivery_id) = 26),
    webhook_id       TEXT NOT NULL REFERENCES webhooks(webhook_id) ON DELETE CASCADE,
    event            TEXT NOT NULL,
    payload          TEXT NOT NULL DEFAULT '{}',
    status           TEXT NOT NULL DEFAULT 'pending',
    attempts         INTEGER NOT NULL DEFAULT 0,
    last_status_code INTEGER,
    last_error       TEXT NOT NULL DEFAULT '',
    next_retry_at    TEXT,
    created_at       TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at     TEXT
);
CREATE INDEX IF NOT EXISTS idx_wd_webhook ON webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS idx_wd_status ON webhook_deliveries(status);
CREATE INDEX IF NOT EXISTS idx_wd_retry ON webhook_deliveries(next_retry_at) WHERE status = 'retrying';
