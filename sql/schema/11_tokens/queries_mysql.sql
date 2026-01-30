-- name: DropTokenTable :exec
DROP TABLE tokens;

-- name: CreateTokenTable :exec
CREATE TABLE IF NOT EXISTS tokens (
    id VARCHAR(26) PRIMARY KEY NOT NULL,
    user_id VARCHAR(26) NOT NULL,
    token_type VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    expires_at TIMESTAMP DEFAULT '0000-00-00 00:00:00' NOT NULL,
    revoked TINYINT(1) DEFAULT 0 NOT NULL,
    CONSTRAINT token
        UNIQUE (token),
    CONSTRAINT fk_tokens_users
        FOREIGN KEY (user_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

-- name: CountToken :one
SELECT COUNT(*)
FROM tokens;

-- name: GetToken :one
SELECT * FROM tokens
WHERE id = ? LIMIT 1;


-- name: GetTokenByUserId :many
SELECT * FROM tokens
WHERE user_id = ?;

-- name: ListToken :many
SELECT * FROM tokens;

-- name: CreateToken :exec
INSERT INTO tokens (
    id,
    user_id,
    token_type,
    token,
    issued_at,
    expires_at,
    revoked
) VALUES(
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: GetLastToken :one
 SELECT * FROM tokens WHERE id = LAST_INSERT_ID();

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

-- name: GetTokenByTokenValue :one
SELECT * FROM tokens
WHERE token = ? LIMIT 1;
