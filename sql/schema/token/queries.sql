-- name: GetToken :one
SELECT * FROM token
WHERE id = ? LIMIT 1;

-- name: CountTokens :one
SELECT COUNT(*)
FROM token;

-- name: GetTokenByUserId :one
SELECT * FROM token
WHERE user_id = ? LIMIT 1;

-- name: CreateToken :one
INSERT INTO token (
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
UPDATE token
set token = ?,
issued_at = ?,
expires_at= ?,
revoked = ?
WHERE id = ?;

-- name: DeleteToken :exec
DELETE FROM token
WHERE id = ?;
