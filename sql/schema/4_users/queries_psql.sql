-- name: DropUserTable :exec
DROP TABLE users;

-- name: CreateUserTable :exec
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL
        PRIMARY KEY,
    username TEXT NOT NULL
        UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role INTEGER NOT NULL DEFAULT 4
        CONSTRAINT fk_users_role
            REFERENCES roles
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- name: CreateUsersEmailIndex :exec
CREATE INDEX idx_users_email 
    ON users (email);

-- name: CountUser :one
SELECT COUNT(*)
FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE user_id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserId :one
SELECT user_id FROM users
WHERE email = $1 LIMIT 1;

-- name: ListUser :many
SELECT * FROM users 
ORDER BY user_id ;

-- name: CreateUser :one
INSERT INTO users (
    username, 
    name, 
    email, 
    hash, 
    role,
    date_created, 
    date_modified
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

-- name: UpdateUser :exec
UPDATE users
SET username = $1,
    name = $2,
    email = $3,
    hash = $4,
    role = $5,
    date_created = $6,
    date_modified = $7
WHERE user_id = $8;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = $1;
