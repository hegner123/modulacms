-- name: CreateSessionTable :exec
CREATE TABLE sessions (
    session_id   INTEGER NOT NULL AUTO_INCREMENT,
    user_id      INTEGER NOT NULL, 
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at   TIMESTAMP,
    last_access  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address   VARCHAR(45),
    user_agent   TEXT,
    session_data TEXT,
    CONSTRAINT fk_sessions_user_id FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- name: GetSession :one
SELECT * FROM sessions
WHERE session_id = ? LIMIT 1;

-- name: CountSessions :one
SELECT COUNT(*)
FROM sessions;

-- name: GetSessionsByUserId :many
SELECT * FROM sessions
WHERE session_id = ?;

-- name: ListSessions :many
SELECT * FROM sessions;

-- name: CreateSession :exec
INSERT INTO sessions (
    user_id,
    created_at,
    expires_at,
    last_access,
    ip_address,
    user_agent,
    session_data
    ) VALUES( 
    ?,?,?,?,?,?,?
    );

-- name: GetLastSession :one
 SELECT * FROM sessions WHERE session_id = LAST_INSERT_ID();

-- name: UpdateSession :exec
UPDATE sessions
    SET user_id=?,
    created_at=?,
    expires_at=?,
    last_access=?,
    ip_address=?,
    user_agent=?,
    session_data=?
WHERE session_id = ?;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE session_id = ?;
