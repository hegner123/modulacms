CREATE TABLE sessions (
    session_id   INTEGER PRIMARY KEY, 
    user_id      INTEGER NOT NULL, 
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at   TIMESTAMP,
    last_access  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address   TEXT,
    user_agent   TEXT,
    session_data TEXT,
    CONSTRAINT fk_sessions_user_id FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);


