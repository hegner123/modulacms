CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id VARCHAR(26) PRIMARY KEY NOT NULL,
    user_id VARCHAR(26) NOT NULL,
    oauth_provider VARCHAR(255) NOT NULL,
    oauth_provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expires_at TIMESTAMP NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT user_oauth_ibfk_1
        FOREIGN KEY (user_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_user_oauth_user ON user_oauth(user_id);
