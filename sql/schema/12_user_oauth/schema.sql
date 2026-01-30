CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(user_oauth_id) = 26),
    user_id TEXT NOT NULL
        REFERENCES users
            ON DELETE CASCADE,
    oauth_provider TEXT NOT NULL,
    oauth_provider_user_id TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expires_at TEXT NOT NULL,
    date_created TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_oauth_user ON user_oauth(user_id);
