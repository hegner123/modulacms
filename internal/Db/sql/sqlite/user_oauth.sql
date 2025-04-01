CREATE TABLE user_oauth (
    user_oauth_id INTEGER
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE CASCADE,
    oauth_provider TEXT NOT NULL,
    oauth_provider_user_id TEXT NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP
);

