CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    oauth_provider VARCHAR(255) NOT NULL,        -- e.g., 'google', 'facebook'
    oauth_provider_user_id VARCHAR(255) NOT NULL,  -- Unique identifier provided by the OAuth provider
    access_token TEXT,                             -- Optional: for making API calls on behalf of the user
    refresh_token TEXT,                            -- Optional: if token refresh is required
    token_expires_at TIMESTAMP NULL,               -- Optional: expiry time for the access token
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

