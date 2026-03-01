-- name: DropWebhookDeliveryTable :exec
DROP TABLE webhook_deliveries;

-- name: CreateWebhookDeliveryTable :exec
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

-- name: CountWebhookDelivery :one
SELECT COUNT(*)
FROM webhook_deliveries;

-- name: GetWebhookDelivery :one
SELECT * FROM webhook_deliveries
WHERE delivery_id = $1 LIMIT 1;

-- name: ListWebhookDeliveries :many
SELECT * FROM webhook_deliveries
ORDER BY created_at DESC;

-- name: ListWebhookDeliveriesByWebhook :many
SELECT * FROM webhook_deliveries
WHERE webhook_id = $1
ORDER BY created_at DESC;

-- name: CreateWebhookDelivery :one
INSERT INTO webhook_deliveries (
    delivery_id,
    webhook_id,
    event,
    payload,
    status,
    attempts,
    created_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
) RETURNING *;

-- name: UpdateWebhookDelivery :exec
UPDATE webhook_deliveries
SET status = $1,
    attempts = $2,
    last_status_code = $3,
    last_error = $4,
    next_retry_at = $5,
    completed_at = $6
WHERE delivery_id = $7;

-- name: DeleteWebhookDelivery :exec
DELETE FROM webhook_deliveries
WHERE delivery_id = $1;

-- name: ListPendingRetries :many
SELECT * FROM webhook_deliveries
WHERE status = 'retrying' AND next_retry_at <= $1
ORDER BY next_retry_at
LIMIT $2;

-- name: UpdateWebhookDeliveryStatus :exec
UPDATE webhook_deliveries
SET status = $1,
    attempts = $2,
    last_status_code = $3,
    last_error = $4,
    next_retry_at = $5,
    completed_at = $6
WHERE delivery_id = $7;

-- name: PruneOldDeliveries :exec
DELETE FROM webhook_deliveries
WHERE status IN ('success', 'failed') AND created_at < $1;
