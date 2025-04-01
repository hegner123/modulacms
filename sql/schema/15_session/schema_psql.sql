CREATE TABLE sessions (
    session_id SERIAL
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        CONSTRAINT fk_sessions_user_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    last_access TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    session_data TEXT
);


