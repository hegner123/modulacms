-- name: DropUserSshKeyTable :exec
DROP TABLE user_ssh_keys;

-- name: CreateUserSshKeyTable :exec
CREATE TABLE IF NOT EXISTS user_ssh_keys (
    ssh_key_id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL,
    public_key TEXT NOT NULL,
    key_type VARCHAR(50) NOT NULL,
    fingerprint VARCHAR(255) NOT NULL UNIQUE,
    label VARCHAR(255),
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- name: CreateUserSshKey :one
INSERT INTO user_ssh_keys (
    ssh_key_id,
    user_id,
    public_key,
    key_type,
    fingerprint,
    label,
    date_created
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserSshKey :one
SELECT * FROM user_ssh_keys
WHERE ssh_key_id = $1
LIMIT 1;

-- name: GetUserSshKeyByFingerprint :one
SELECT * FROM user_ssh_keys
WHERE fingerprint = $1
LIMIT 1;

-- name: GetUserBySSHFingerprint :one
SELECT u.* FROM users u
INNER JOIN user_ssh_keys k ON u.user_id = k.user_id
WHERE k.fingerprint = $1
LIMIT 1;

-- name: ListUserSshKeys :many
SELECT * FROM user_ssh_keys
WHERE user_id = $1
ORDER BY date_created DESC;

-- name: UpdateUserSshKeyLastUsed :exec
UPDATE user_ssh_keys
SET last_used = $1
WHERE ssh_key_id = $2;

-- name: UpdateUserSshKeyLabel :exec
UPDATE user_ssh_keys
SET label = $1
WHERE ssh_key_id = $2;

-- name: DeleteUserSshKey :exec
DELETE FROM user_ssh_keys
WHERE ssh_key_id = $1;

-- name: CountUserSshKeys :one
SELECT COUNT(*) FROM user_ssh_keys;
