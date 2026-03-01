-- name: DropWebhookTable :exec
DROP TABLE webhooks;

-- name: CreateWebhookTable :exec
CREATE TABLE IF NOT EXISTS webhooks (
    webhook_id    TEXT PRIMARY KEY NOT NULL,
    name          TEXT NOT NULL,
    url           TEXT NOT NULL,
    secret        TEXT NOT NULL DEFAULT '',
    events        TEXT NOT NULL DEFAULT '[]',
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    headers       TEXT NOT NULL DEFAULT '{}',
    author_id     TEXT NOT NULL REFERENCES users(user_id),
    date_created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- name: CountWebhook :one
SELECT COUNT(*)
FROM webhooks;

-- name: GetWebhook :one
SELECT * FROM webhooks
WHERE webhook_id = $1 LIMIT 1;

-- name: ListWebhooks :many
SELECT * FROM webhooks
ORDER BY date_created DESC;

-- name: ListWebhooksPaginated :many
SELECT * FROM webhooks
ORDER BY date_created DESC
LIMIT $1 OFFSET $2;

-- name: ListActiveWebhooks :many
SELECT * FROM webhooks
WHERE is_active = TRUE
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
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
) RETURNING *;

-- name: UpdateWebhook :exec
UPDATE webhooks
SET name = $1,
    url = $2,
    secret = $3,
    events = $4,
    is_active = $5,
    headers = $6,
    date_modified = $7
WHERE webhook_id = $8;

-- name: DeleteWebhook :exec
DELETE FROM webhooks
WHERE webhook_id = $1;
