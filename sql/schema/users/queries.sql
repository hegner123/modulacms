
-- name: GetUser :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users 
ORDER BY id ;

-- name: CreateUser :one
INSERT INTO users (
    datecreated,
    datemodified,
    username,
    name,
    email ,
    hash,
    role
) VALUES (
?,?,?,?,?,?,?
)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users
set datecreated = ?,
    datemodified = ?,
    username = ?,
    name = ?,
    email = ?,
    hash = ?,
    role = ?
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;
