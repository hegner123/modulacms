-- name: GetUser :one
SELECT * FROM user
WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM user
WHERE email = ? LIMIT 1;

-- name: GetUserId :one
SELECT id FROM user
WHERE email = ? LIMIT 1;

-- name: ListUser :many
SELECT * FROM user 
ORDER BY id ;

-- name: CreateUser :one
INSERT INTO user (
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
UPDATE user
set datecreated = ?,
    datemodified = ?,
    username = ?,
    name = ?,
    email = ?,
    hash = ?,
    role = ?
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM user
WHERE id = ?;
