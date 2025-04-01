-- name: DropPermissionTable :exec
DROP TABLE permissions;

-- name: CreatePermissionTable :exec
CREATE TABLE IF NOT EXISTS permissions (
    permission_id INTEGER
        PRIMARY KEY,
    table_id INTEGER NOT NULL,
    mode INTEGER NOT NULL,
    label TEXT NOT NULL
);

-- name: GetPermission :one
SELECT * FROM permissions 
WHERE permission_id = ? LIMIT 1;

-- name: CountPermission :one
SELECT COUNT(*)
FROM permissions;

-- name: ListPermission :many
SELECT * FROM permissions 
ORDER BY table_id;

-- name: CreatePermission :one
INSERT INTO permissions(
    table_id,
    mode,
    label
) VALUES (
    ?,
    ?,
    ?
)
RETURNING *;

-- name: UpdatePermission :exec
UPDATE permissions
SET table_id=?,
    mode=?,
    label=?
WHERE permission_id = ?;

-- name: DeletePermission :exec
DELETE FROM permissions 
WHERE permission_id = ?;
