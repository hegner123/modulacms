-- name: DropSessionTable :exec
DROP TABLE sessions;

-- name: CreateSessionTable :exec
CREATE TABLE IF NOT EXISTS sessions (
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

-- name: GetSession :one
SELECT * FROM sessions
WHERE session_id = $1 LIMIT 1;

-- name: CountSession :one
SELECT COUNT(*)
FROM sessions;

-- name: GetSessionByUserId :one
SELECT * FROM sessions
WHERE user_id = $1
ORDER BY session_id DESC
LIMIT 1;

-- name: ListSession :many
SELECT * FROM sessions;

-- name: CreateSession :one
INSERT INTO sessions (
    session_id,
    user_id,
    created_at,
    expires_at,
    last_access,
    ip_address,
    user_agent,
    session_data
) VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
) RETURNING *;

-- name: UpdateSession :exec
UPDATE sessions
    SET user_id=$1,
    created_at=$2,
    expires_at=$3,
    last_access=$4,
    ip_address=$5,
    user_agent=$6,
    session_data=$7
WHERE session_id = $8;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE session_id = $1;
