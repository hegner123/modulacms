-- name: CreatePermissionTable :exec
CREATE TABLE IF NOT EXISTS permissions (
    permission_id INT PRIMARY KEY,
    table_id INT NOT NULL,
    mode INT NOT NULL,
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
  ?,?,?
)
RETURNING *;

-- name: UpdatePermission :exec
UPDATE permissions
set table_id=?,
    mode=?,
    label=?
WHERE permission_id = ?;

-- name: DeletePermission :exec
DELETE FROM permissions 
WHERE permission_id = ?;
