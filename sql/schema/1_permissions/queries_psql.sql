-- name: DropPermissionTable :exec
DROP TABLE permissions;

-- name: CreatePermissionTable :exec
CREATE TABLE IF NOT EXISTS permissions (
    permission_id TEXT PRIMARY KEY NOT NULL,
    table_id TEXT NOT NULL,
    mode INTEGER NOT NULL,
    label TEXT NOT NULL
);
-- name: GetPermission :one
SELECT * FROM permissions 
WHERE permission_id = $1 LIMIT 1;

-- name: CountPermission :one
SELECT COUNT(*)
FROM permissions;

-- name: ListPermission :many
SELECT * FROM permissions 
ORDER BY table_id;

-- name: CreatePermission :one
INSERT INTO permissions(
    permission_id,
    table_id,
    mode,
    label
) VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: UpdatePermission :exec
UPDATE permissions
SET table_id=$1,
    mode=$2,
    label=$3
WHERE permission_id = $4;

-- name: DeletePermission :exec
DELETE FROM permissions 
WHERE permission_id = $1;
