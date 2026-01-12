CREATE TABLE sessions (
    session_id INTEGER
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE CASCADE,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    expires_at TEXT,
    last_access TEXT DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    session_data TEXT
);


