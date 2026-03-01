CREATE TABLE IF NOT EXISTS webhook_deliveries (
    delivery_id      VARCHAR(26) PRIMARY KEY NOT NULL,
    webhook_id       VARCHAR(26) NOT NULL,
    event            VARCHAR(255) NOT NULL,
    payload          MEDIUMTEXT NOT NULL,
    status           VARCHAR(50) NOT NULL DEFAULT 'pending',
    attempts         INT NOT NULL DEFAULT 0,
    last_status_code INT,
    last_error       TEXT NOT NULL,
    next_retry_at    TIMESTAMP NULL,
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    completed_at     TIMESTAMP NULL,
    CONSTRAINT fk_wd_webhook FOREIGN KEY (webhook_id) REFERENCES webhooks(webhook_id) ON DELETE CASCADE
);
CREATE INDEX idx_wd_webhook ON webhook_deliveries(webhook_id);
CREATE INDEX idx_wd_status ON webhook_deliveries(status);
CREATE INDEX idx_wd_retry ON webhook_deliveries(next_retry_at);
