-- name: CreateSessionTable :exec
CREATE TABLE sessions (
    session_id   INTEGER PRIMARY KEY,
    user_id      INTEGER NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE CASCADE,
    created_at   TEXT DEFAULT CURRENT_TIMESTAMP,
    expires_at   TEXT,
    last_access  TEXT DEFAULT CURRENT_TIMESTAMP,
    ip_address   TEXT,
    user_agent   TEXT,
    session_data TEXT 
);

-- name: GetSession :one
SELECT * FROM sessions
WHERE session_id = ? LIMIT 1;

-- name: CountSessions :one
SELECT COUNT(*)
FROM sessions;

-- name: GetSessionsByUserId :many
SELECT * FROM sessions
WHERE user_id = ?;

-- name: ListSessions :many
SELECT * FROM sessions;

-- name: CreateSession :one
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
    ) RETURNING *;

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
