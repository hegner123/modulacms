CREATE TABLE sessions (
    session_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(session_id) = 26),
    user_id TEXT NOT NULL
        REFERENCES users
            ON DELETE CASCADE,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    expires_at TEXT,
    last_access TEXT DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    session_data TEXT
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);

