-- name: DropWebhookTable :exec
DROP TABLE webhooks;

-- name: CreateWebhookTable :exec
CREATE TABLE IF NOT EXISTS webhooks (
    webhook_id    VARCHAR(26) PRIMARY KEY NOT NULL,
    name          VARCHAR(255) NOT NULL,
    url           TEXT NOT NULL,
    secret        TEXT NOT NULL,
    events        TEXT NOT NULL,
    is_active     TINYINT NOT NULL DEFAULT 1,
    headers       TEXT NOT NULL,
    author_id     VARCHAR(26) NOT NULL,
    date_created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_webhooks_author FOREIGN KEY (author_id) REFERENCES users(user_id)
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

-- name: CreateWebhook :exec
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
);

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
