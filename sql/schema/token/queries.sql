-- name: CreateTokenTable :exec
CREATE TABLE IF NOT EXISTS tokens (
    id INTEGER
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL
        UNIQUE,
    issued_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    revoked BOOLEAN DEFAULT 0
);
-- name: GetToken :one
SELECT * FROM tokens
WHERE id = ? LIMIT 1;

-- name: CountTokens :one
SELECT COUNT(*)
FROM tokens;

-- name: GetTokensByUserId :many
SELECT * FROM tokens
WHERE user_id = ?;

-- name: ListTokens :many
SELECT * FROM tokens;

-- name: CreateToken :one
INSERT INTO tokens (
    user_id,
    token_type,
    token,
    issued_at,
    expires_at,
    revoked
    ) VALUES( 
    ?,?,?,?,?,?
    ) RETURNING *;

-- name: UpdateToken :exec
UPDATE tokens
set token = ?,
issued_at = ?,
expires_at = ?,
revoked = ?
WHERE id = ?;

-- name: DeleteToken :exec
DELETE FROM tokens
WHERE id = ?;
