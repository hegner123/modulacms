-- name: GetRole :one
SELECT * FROM roles
WHERE role_id = ? LIMIT 1;

-- name: ListRole :many
SELECT * FROM roles 
ORDER BY role_id ;

-- name: CreateRole :one
INSERT INTO roles (
    label,
    permissions
) VALUES (
?,?
)
RETURNING *;

-- name: UpdateRole :exec
UPDATE roles
set label=?,
    permissions=?
WHERE role_id = ?;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE role_id = ?;
