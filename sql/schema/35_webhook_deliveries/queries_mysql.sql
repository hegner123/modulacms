-- name: DropWebhookDeliveryTable :exec
DROP TABLE webhook_deliveries;

-- name: CreateWebhookDeliveryTable :exec
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

-- name: CountWebhookDelivery :one
SELECT COUNT(*)
FROM webhook_deliveries;

-- name: GetWebhookDelivery :one
SELECT * FROM webhook_deliveries
WHERE delivery_id = ? LIMIT 1;

-- name: ListWebhookDeliveries :many
SELECT * FROM webhook_deliveries
ORDER BY created_at DESC;

-- name: ListWebhookDeliveriesByWebhook :many
SELECT * FROM webhook_deliveries
WHERE webhook_id = ?
ORDER BY created_at DESC;

-- name: CreateWebhookDelivery :exec
INSERT INTO webhook_deliveries (
    delivery_id,
    webhook_id,
    event,
    payload,
    status,
    attempts,
    created_at
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: UpdateWebhookDelivery :exec
UPDATE webhook_deliveries
SET status = ?,
    attempts = ?,
    last_status_code = ?,
    last_error = ?,
    next_retry_at = ?,
    completed_at = ?
WHERE delivery_id = ?;

-- name: DeleteWebhookDelivery :exec
DELETE FROM webhook_deliveries
WHERE delivery_id = ?;

-- name: ListPendingRetries :many
SELECT * FROM webhook_deliveries
WHERE status = 'retrying' AND next_retry_at <= ?
ORDER BY next_retry_at
LIMIT ?;

-- name: UpdateWebhookDeliveryStatus :exec
UPDATE webhook_deliveries
SET status = ?,
    attempts = ?,
    last_status_code = ?,
    last_error = ?,
    next_retry_at = ?,
    completed_at = ?
WHERE delivery_id = ?;

-- name: PruneOldDeliveries :exec
DELETE FROM webhook_deliveries
WHERE status IN ('success', 'failed') AND created_at < ?;
