CREATE TABLE IF NOT EXISTS webhook_deliveries (
    delivery_id      TEXT PRIMARY KEY NOT NULL,
    webhook_id       TEXT NOT NULL REFERENCES webhooks(webhook_id) ON DELETE CASCADE,
    event            TEXT NOT NULL,
    payload          TEXT NOT NULL DEFAULT '{}',
    status           TEXT NOT NULL DEFAULT 'pending',
    attempts         INTEGER NOT NULL DEFAULT 0,
    last_status_code INTEGER,
    last_error       TEXT NOT NULL DEFAULT '',
    next_retry_at    TIMESTAMP,
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    completed_at     TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_wd_webhook ON webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS idx_wd_status ON webhook_deliveries(status);
CREATE INDEX IF NOT EXISTS idx_wd_retry ON webhook_deliveries(next_retry_at) WHERE status = 'retrying';
