-- name: DropWebhookDeliveryTable :exec
DROP TABLE IF EXISTS webhook_deliveries;

-- name: CreateWebhookDeliveryTable :exec
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
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING *;

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
