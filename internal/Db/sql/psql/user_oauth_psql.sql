CREATE TABLE user_oauth (
    user_oauth_id SERIAL
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE CASCADE,
    oauth_provider VARCHAR(255) NOT NULL,
    oauth_provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE user_oauth
    OWNER TO modula_u;

