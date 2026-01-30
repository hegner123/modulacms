-- name: DropUserTable :exec
DROP TABLE users;

-- name: CreateUserTable :exec
CREATE TABLE IF NOT EXISTS users (
    user_id TEXT PRIMARY KEY NOT NULL CHECK (length(user_id) = 26),
    username TEXT NOT NULL
        UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role TEXT NOT NULL
        REFERENCES roles
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- name: CreateUsersEmailIndex :exec
CREATE UNIQUE INDEX users_email_uindex
    ON users (email);

-- name: GetUser :one
SELECT * FROM users
WHERE user_id = ? LIMIT 1;

-- name: CountUser :one
SELECT COUNT(*)
FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ? LIMIT 1;

-- name: GetUserId :one
SELECT user_id FROM users
WHERE email = ? LIMIT 1;

-- name: ListUser :many
SELECT * FROM users 
ORDER BY user_id ;

-- name: CreateUser :one
INSERT INTO users (
    user_id,
    username,
    name,
    email,
    hash,
    role,
    date_created,
    date_modified
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users
SET username = ?, 
    name = ?, 
    email = ?, 
    hash = ?, 
    role = ?,
    date_created = ?, 
    date_modified = ?
WHERE user_id = ?;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = ?;
