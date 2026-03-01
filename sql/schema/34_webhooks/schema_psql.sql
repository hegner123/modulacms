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
CREATE INDEX IF NOT EXISTS idx_webhooks_active ON webhooks(is_active);
