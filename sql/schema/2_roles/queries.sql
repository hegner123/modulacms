-- name: DropRoleTable :exec
DROP TABLE roles;

-- name: CreateRoleTable :exec
CREATE TABLE IF NOT EXISTS roles (
    role_id TEXT PRIMARY KEY NOT NULL CHECK (length(role_id) = 26),
    label TEXT NOT NULL
        UNIQUE,
    system_protected INTEGER NOT NULL DEFAULT 0
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
    role_id,
    label,
    system_protected
) VALUES (
    ?,
    ?,
    ?
)
RETURNING *;

-- name: UpdateRole :exec
UPDATE roles
SET label = ?,
    system_protected = ?
WHERE role_id = ?;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE role_id = ?;
