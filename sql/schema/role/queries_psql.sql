-- name: GetRole :one
SELECT * FROM roles
WHERE role_id = $1;

-- name: ListRole :many
SELECT * FROM roles 
ORDER BY role_id;

-- name: CreateRole :one
INSERT INTO roles (
    label,
    permissions
) VALUES (
    $1, $2
)
RETURNING *;

-- name: UpdateRole :exec
UPDATE roles
SET label = $1,
    permissions = $2
WHERE role_id = $3;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE role_id = $1;

