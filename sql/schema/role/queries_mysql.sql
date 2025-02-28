-- name: GetRole :one
SELECT * FROM roles
WHERE role_id = ? LIMIT 1;

-- name: ListRole :many
SELECT * FROM roles;

-- name: CreateRole :exec
INSERT INTO roles (label, permissions) VALUES (?,?);

-- name: GetLastRole :one
SELECT * FROM roles WHERE role_id = LAST_INSERT_ID();

-- name: UpdateRole :exec
UPDATE roles
set label=?,
    permissions=?
WHERE role_id = ?;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE role_id = ?;
