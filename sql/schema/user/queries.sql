-- name: GetUser :one
SELECT * FROM user
WHERE user_id = ? LIMIT 1;

-- name: CountUsers :one
SELECT COUNT(*)
FROM user;

-- name: GetUserByEmail :one
SELECT * FROM user
WHERE email = ? LIMIT 1;

-- name: GetUserId :one
SELECT user_id FROM user
WHERE email = ? LIMIT 1;

-- name: ListUser :many
SELECT * FROM user 
ORDER BY user_id ;

-- name: CreateUser :one
INSERT INTO user (
    date_created,
    date_modified,
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
set date_created = ?,
    date_modified = ?,
    username = ?,
    name = ?,
    email = ?,
    hash = ?,
    role = ?
WHERE user_id = ?;

-- name: DeleteUser :exec
DELETE FROM user
WHERE user_id = ?;
