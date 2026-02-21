-- name: DropSessionTable :exec
DROP TABLE sessions;

-- name: CreateSessionTable :exec
CREATE TABLE IF NOT EXISTS sessions (
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

-- name: CountSession :one
SELECT COUNT(*)
FROM sessions;

-- name: GetSession :one
SELECT * FROM sessions
WHERE session_id = ? LIMIT 1;

-- name: GetSessionByUserId :one
SELECT * FROM sessions
WHERE user_id = ?
ORDER BY session_id DESC
LIMIT 1;

-- name: ListSession :many
SELECT * FROM sessions;

-- name: CreateSession :one
INSERT INTO sessions (
    session_id,
    user_id,
    date_created,
    expires_at,
    last_access,
    ip_address,
    user_agent,
    session_data
) VALUES(
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING *;

-- name: UpdateSession :exec
UPDATE sessions
    SET user_id=?,
    date_created=?,
    expires_at=?,
    last_access=?,
    ip_address=?,
    user_agent=?,
    session_data=?
WHERE session_id = ?;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE session_id = ?;
