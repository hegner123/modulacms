CREATE TABLE token (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    issued_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    revoked BOOLEAN DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES user (id)
);

