-- name: CreateUserTable :exec
CREATE TABLE IF NOT EXISTS users (
    user_id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role INTEGER NOT NULL,
        references role
            on update cascade on delete set NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
-- name: GetUser :one
SELECT * FROM users
WHERE user_id = ? LIMIT 1;

-- name: CountUsers :one
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
    date_created,
    date_modified,
    username,
    name,
    email,
    hash,
    role
) VALUES (
?,?,?,?,?,?,?
)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users
set date_created = ?,
    date_modified = ?,
    username = ?,
    name = ?,
    email = ?,
    hash = ?,
    role = ?
WHERE user_id = ?;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = ?;
