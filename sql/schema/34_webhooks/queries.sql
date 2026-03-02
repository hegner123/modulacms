-- name: DropWebhookTable :exec
DROP TABLE IF EXISTS webhooks;

-- name: CreateWebhookTable :exec
CREATE TABLE IF NOT EXISTS webhooks (
    webhook_id    TEXT PRIMARY KEY NOT NULL CHECK (length(webhook_id) = 26),
    name          TEXT NOT NULL,
    url           TEXT NOT NULL,
    secret        TEXT NOT NULL DEFAULT '',
    events        TEXT NOT NULL DEFAULT '[]',
    is_active     INTEGER NOT NULL DEFAULT 1,
    headers       TEXT NOT NULL DEFAULT '{}',
    author_id     TEXT NOT NULL REFERENCES users(user_id),
    date_created  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- name: CountWebhook :one
SELECT COUNT(*)
FROM webhooks;

-- name: GetWebhook :one
SELECT * FROM webhooks
WHERE webhook_id = ? LIMIT 1;

-- name: ListWebhooks :many
SELECT * FROM webhooks
ORDER BY date_created DESC;

-- name: ListWebhooksPaginated :many
SELECT * FROM webhooks
ORDER BY date_created DESC
LIMIT ? OFFSET ?;

-- name: ListActiveWebhooks :many
SELECT * FROM webhooks
WHERE is_active = 1
ORDER BY date_created DESC;

-- name: CreateWebhook :one
INSERT INTO webhooks (
    webhook_id,
    name,
    url,
    secret,
    events,
    is_active,
    headers,
    author_id,
    date_created,
    date_modified
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING *;

-- name: UpdateWebhook :exec
UPDATE webhooks
SET name = ?,
    url = ?,
    secret = ?,
    events = ?,
    is_active = ?,
    headers = ?,
    date_modified = ?
WHERE webhook_id = ?;

-- name: DeleteWebhook :exec
DELETE FROM webhooks
WHERE webhook_id = ?;
