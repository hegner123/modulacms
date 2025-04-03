-- name: DropUserOauthTable :exec
DROP TABLE user_oauth;

-- name: CreateUserOauthTable :exec
CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id SERIAL
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE CASCADE,
    oauth_provider VARCHAR(255) NOT NULL,
    oauth_provider_user_id VARCHAR(255) NOT NULL,
    access_token  TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expires_at TIMESTAMP NOT NULL,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- name: CountUserOauths :one
SELECT COUNT(*)
FROM user_oauth;

-- name: GetUserOauth :one
SELECT *
FROM user_oauth
WHERE user_oauth_id = $1
LIMIT 1;

-- name: GetUserOauthByEmail :one
SELECT uo.*
FROM user_oauth uo
JOIN users u ON uo.user_id = u.user_id
WHERE u.email = $1
LIMIT 1;

-- name: GetUserOauthId :one
SELECT uo.user_id
FROM user_oauth uo
JOIN users u ON uo.user_id = u.user_id
WHERE u.email = $1
LIMIT 1;

-- name: ListUserOauth :many
SELECT *
FROM user_oauth
ORDER BY user_oauth_id;

-- name: CreateUserOauth :one
INSERT INTO user_oauth (
    user_id,
    oauth_provider,
    oauth_provider_user_id,
    access_token,
    refresh_token,
    token_expires_at,
    date_created
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;

-- name: UpdateUserOauth :exec
UPDATE user_oauth
SET access_token = $1,
    refresh_token = $2,
    token_expires_at = $3
WHERE user_oauth_id = $4;

-- name: DeleteUserOauth :exec
DELETE FROM user_oauth
WHERE user_oauth_id = $1;

