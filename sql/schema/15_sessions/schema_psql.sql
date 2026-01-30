CREATE TABLE sessions (
    session_id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL
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

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);

