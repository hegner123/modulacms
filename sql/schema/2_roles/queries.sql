-- name: DropRoleTable :exec
DROP TABLE roles;

-- name: CreateRoleTable :exec
CREATE TABLE IF NOT EXISTS roles (
    role_id INTEGER
        PRIMARY KEY,
    label TEXT NOT NULL
        UNIQUE,
    permissions TEXT NOT NULL
        UNIQUE
);

-- name: GetRole :one
SELECT * 
FROM roles
WHERE role_id = ? 
LIMIT 1;

-- name: CountRole :one
SELECT COUNT(*)
FROM roles;

-- name: ListRole :many
SELECT * 
FROM roles 
ORDER BY role_id;

-- name: CreateRole :one
INSERT INTO roles (
    label,
    permissions
) VALUES (
    ?,
    ?
)
RETURNING *;

-- name: UpdateRole :exec
UPDATE roles
SET label = ?,
    permissions = ?
WHERE role_id = ?;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE role_id = ?;
