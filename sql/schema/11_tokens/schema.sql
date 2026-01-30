CREATE TABLE IF NOT EXISTS tokens (
    id TEXT
        PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    user_id TEXT NOT NULL
        REFERENCES users
            ON DELETE CASCADE,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL
        UNIQUE,
    issued_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_tokens_user ON tokens(user_id);

