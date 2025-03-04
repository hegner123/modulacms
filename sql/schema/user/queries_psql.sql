-- name: CreateUserTable :exec
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role INTEGER,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_users_role FOREIGN KEY (role)
        REFERENCES roles(role_id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

-- name: GetUser :one
SELECT * FROM users
WHERE user_id = $1 LIMIT 1;

-- name: CountUsers :one
SELECT COUNT(*)
FROM users;

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
    date_created,
    date_modified,
    username,
    name,
    email,
    hash,
    role
) VALUES (
$1,$2,$3,$4,$5,$6,$7
)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users
set date_created = $1,
    date_modified = $2,
    username = $3,
    name = $4,
    email = $5,
    hash = $6,
    role = $7
WHERE user_id = $8;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = $1;
