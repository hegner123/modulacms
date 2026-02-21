-- name: DropSessionTable :exec
DROP TABLE sessions;

-- name: CreateSessionTable :exec
CREATE TABLE IF NOT EXISTS sessions (
    session_id VARCHAR(26) PRIMARY KEY NOT NULL,
    user_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expires_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_access TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ip_address VARCHAR(45) NULL,
    user_agent TEXT NULL,
    session_data TEXT NULL,
    CONSTRAINT fk_sessions_user_id
        FOREIGN KEY (user_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE
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

-- name: CreateSession :exec
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
);

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
