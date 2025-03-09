CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    oauth_provider TEXT NOT NULL,        -- e.g., 'google', 'facebook'
    oauth_provider_user_id TEXT NOT NULL,  -- Unique identifier provided by the OAuth provider
    access_token TEXT,                     -- Optional: for making API calls on behalf of the user
    refresh_token TEXT,                    -- Optional: if token refresh is required
    token_expires_at TEXT,                 -- Optional: expiry time for the access token
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);
