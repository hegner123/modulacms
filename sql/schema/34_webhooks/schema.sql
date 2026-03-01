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
CREATE INDEX IF NOT EXISTS idx_webhooks_active ON webhooks(is_active);
