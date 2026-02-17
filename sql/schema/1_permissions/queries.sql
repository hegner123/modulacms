-- name: DropPermissionTable :exec
DROP TABLE permissions;

-- name: CreatePermissionTable :exec
CREATE TABLE IF NOT EXISTS permissions (
    permission_id TEXT PRIMARY KEY NOT NULL CHECK (length(permission_id) = 26),
    label TEXT NOT NULL UNIQUE,
    system_protected INTEGER NOT NULL DEFAULT 0
);

-- name: GetPermission :one
SELECT * FROM permissions
WHERE permission_id = ? LIMIT 1;

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
    ?,
    ?,
    ?
)
RETURNING *;

-- name: UpdatePermission :exec
UPDATE permissions
SET label=?,
    system_protected=?
WHERE permission_id = ?;

-- name: DeletePermission :exec
DELETE FROM permissions
WHERE permission_id = ?;
