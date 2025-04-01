CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id INT AUTO_INCREMENT
        PRIMARY KEY,
    user_id INT NOT NULL,
    oauth_provider VARCHAR(255) NOT NULL,
    oauth_provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT NULL,
    refresh_token TEXT NULL,
    token_expires_at TIMESTAMP NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP() NOT NULL,
    CONSTRAINT user_oauth_ibfk_1
        FOREIGN KEY (user_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);
