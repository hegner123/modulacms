-- name: CreateTokenTable :exec
CREATE TABLE IF NOT EXISTS tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT false,
    CONSTRAINT fk_tokens_users FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

-- name: GetToken :one
SELECT * FROM tokens
WHERE id = $1 LIMIT 1;

-- name: CountToken :one
SELECT COUNT(*)
FROM tokens;

-- name: GetTokenByUserId :many
SELECT * FROM tokens
WHERE user_id = $1;

-- name: ListToken :many
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
    $1,$2,$3,$4,$5,$6
    ) RETURNING *;

-- name: UpdateToken :exec
UPDATE tokens
set token = $1,
issued_at = $2,
expires_at = $3,
revoked = $4
WHERE id = $5;

-- name: DeleteToken :exec
DELETE FROM tokens
WHERE id = $1;
