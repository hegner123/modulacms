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
CREATE INDEX idx_webhooks_active ON webhooks(is_active);
