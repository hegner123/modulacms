CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id SERIAL
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE CASCADE,
    oauth_provider VARCHAR(255) NOT NULL,
    oauth_provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expires_at TIMESTAMP NOT NULL,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_oauth_user ON user_oauth(user_id);
