-- name: DropTokenTable :exec
DROP TABLE tokens;

-- name: CreateTokenTable :exec
CREATE TABLE IF NOT EXISTS tokens (
    id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT fk_tokens_users FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE
);

-- name: CountToken :one
SELECT COUNT(*)
FROM tokens;

-- name: GetToken :one
SELECT * FROM tokens
WHERE id = $1 LIMIT 1;

-- name: GetTokenByUserId :many
SELECT * FROM tokens
WHERE user_id = $1;

-- name: ListToken :many
SELECT * FROM tokens;

-- name: CreateToken :one
INSERT INTO tokens (
    id,
    user_id,
    token_type,
    token,
    issued_at,
    expires_at,
    revoked
    ) VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
) RETURNING *;

-- name: UpdateToken :exec
UPDATE tokens
SET token = $1,
    issued_at = $2,
    expires_at = $3,
    revoked = $4
WHERE id = $5;

-- name: DeleteToken :exec
DELETE FROM tokens
WHERE id = $1;

-- name: GetTokenByTokenValue :one
SELECT * FROM tokens
WHERE token = $1 LIMIT 1;
