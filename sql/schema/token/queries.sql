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
