-- name: CreateUserOauthTable :exec
CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    oauth_provider TEXT NOT NULL,        -- e.g., 'google', 'facebook'
    oauth_provider_user_id TEXT NOT NULL,  -- Unique identifier provided by the OAuth provider
    access_token TEXT,                     -- Optional: for making API calls on behalf of the user
    refresh_token TEXT,                    -- Optional: if token refresh is required
    token_expires_at TEXT,                 -- Optional: expiry time for the access token
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

-- name: GetUserOauth :one
SELECT *
FROM user_oauth
WHERE user_oauth_id = ?
LIMIT 1;

-- name: CountUserOauths :one
SELECT COUNT(*)
FROM user_oauth;

-- name: GetUserOauthByEmail :one
SELECT uo.*
FROM user_oauth uo
JOIN users u ON uo.user_id = u.user_id
WHERE u.email = ?
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
    ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpdateUserOauth :exec
UPDATE user_oauth
SET access_token = ?,
    refresh_token = ?,
    token_expires_at = ?
WHERE user_oauth_id = ?;

-- name: DeleteUserOauth :exec
DELETE FROM user_oauth
WHERE user_oauth_id = ?;

