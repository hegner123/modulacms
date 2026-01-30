-- name: DropPermissionTable :exec
DROP TABLE permissions;

-- name: CreatePermissionTable :exec
CREATE TABLE IF NOT EXISTS permissions (
    permission_id VARCHAR(26) PRIMARY KEY NOT NULL,
    table_id VARCHAR(26) NOT NULL,
    mode INT NOT NULL,
    label VARCHAR(255) NOT NULL
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

-- name: CreatePermission :exec
INSERT INTO permissions(
    permission_id,
    table_id,
    mode,
    label
) VALUES (
    ?,
    ?,
    ?,
    ?
);

-- name: GetLastPermission :one
SELECT * FROM permissions WHERE permission_id = LAST_INSERT_ID();

-- name: UpdatePermission :exec
UPDATE permissions
set table_id=?,
    mode=?,
    label=?
WHERE permission_id = ?;

-- name: DeletePermission :exec
DELETE FROM permissions 
WHERE permission_id = ?;
