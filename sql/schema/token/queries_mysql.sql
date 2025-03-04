-- name: CreateTokenTable :exec
CREATE TABLE IF NOT EXISTS tokens (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    token_type VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked TINYINT(1) DEFAULT 0,
    CONSTRAINT fk_tokens_users FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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

-- name: CreateToken :exec
INSERT INTO tokens (
    user_id,
    token_type,
    token,
    issued_at,
    expires_at,
    revoked
    ) VALUES( 
    ?,?,?,?,?,?
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
