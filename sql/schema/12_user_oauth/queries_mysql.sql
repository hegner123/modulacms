-- name: DropUserOauthTable :exec
DROP TABLE user_oauth;

-- name: CreateUserOauthTable :exec
CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id INT AUTO_INCREMENT
        PRIMARY KEY,
    user_id INT NOT NULL,
    oauth_provider VARCHAR(255) NOT NULL,
    oauth_provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expires_at TIMESTAMP NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT user_oauth_ibfk_1
        FOREIGN KEY (user_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

-- name: CountUserOauths :one
SELECT COUNT(*)
FROM user_oauth;

-- name: GetUserOauth :one
SELECT *
FROM user_oauth
WHERE user_oauth_id = ?
LIMIT 1;

-- name: GetUserOauthByEmail :one
SELECT uo.*
FROM user_oauth uo
JOIN users u ON uo.user_id = u.user_id
WHERE u.email = ?
LIMIT 1;

-- name: GetUserOauthByUserId :one
SELECT *
FROM user_oauth
WHERE user_id = ?
LIMIT 1;

-- name: GetUserOauthByProviderID :one
SELECT *
FROM user_oauth
WHERE oauth_provider = ? AND oauth_provider_user_id = ?
LIMIT 1;

-- name: GetUserOauthId :one
SELECT uo.user_id
FROM user_oauth uo
JOIN users u ON uo.user_id = u.user_id
WHERE u.email = ?
LIMIT 1;

-- name: ListUserOauth :many
SELECT *
FROM user_oauth
ORDER BY user_oauth_id;

-- name: CreateUserOauth :exec
INSERT INTO user_oauth (
    user_id,
    oauth_provider,
    oauth_provider_user_id,
    access_token,
    refresh_token,
    token_expires_at,
    date_created
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: GetLastUserOauth :one
SELECT *
FROM user_oauth
WHERE user_oauth_id = LAST_INSERT_ID();

-- name: UpdateUserOauth :exec
UPDATE user_oauth
SET access_token = ?,
    refresh_token = ?,
    token_expires_at = ?
WHERE user_oauth_id = ?;

-- name: DeleteUserOauth :exec
DELETE FROM user_oauth
WHERE user_oauth_id = ?;

