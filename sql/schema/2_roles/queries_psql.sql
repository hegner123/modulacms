-- name: DropRoleTable :exec
DROP TABLE roles;

-- name: CreateRoleTable :exec
CREATE TABLE IF NOT EXISTS roles (
    role_id TEXT PRIMARY KEY NOT NULL,
    label TEXT NOT NULL
        UNIQUE,
    permissions jsonb
);

-- name: CreateIDIndex :exec
CREATE INDEX IF NOT EXISTS idx_role_id ON roles(role_id);

-- name: GetRole :one
SELECT * 
FROM roles
WHERE role_id = $1;

-- name: CountRole :one
SELECT COUNT(*)
FROM roles;

-- name: ListRole :many
SELECT * 
FROM roles 
ORDER BY role_id;

-- name: CreateRole :one
INSERT INTO roles (
    role_id,
    label,
    permissions
) VALUES (
    $1,
    $2,
    $3
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
