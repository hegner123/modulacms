-- name: DropPermissionTable :exec
DROP TABLE permissions;

-- name: CreatePermissionTable :exec
CREATE TABLE IF NOT EXISTS permissions (
    permission_id TEXT PRIMARY KEY NOT NULL,
    label TEXT NOT NULL UNIQUE,
    system_protected BOOLEAN NOT NULL DEFAULT FALSE
);
-- name: GetPermission :one
SELECT * FROM permissions
WHERE permission_id = $1 LIMIT 1;

-- name: CountPermission :one
SELECT COUNT(*)
FROM permissions;

-- name: ListPermission :many
SELECT * FROM permissions
ORDER BY label;

-- name: CreatePermission :one
INSERT INTO permissions(
    permission_id,
    label,
    system_protected
) VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: UpdatePermission :exec
UPDATE permissions
SET label=$1,
    system_protected=$2
WHERE permission_id = $3;

-- name: DeletePermission :exec
DELETE FROM permissions
WHERE permission_id = $1;
