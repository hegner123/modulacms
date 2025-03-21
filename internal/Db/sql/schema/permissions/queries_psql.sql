-- name: CreatePermissionTable :exec
CREATE TABLE IF NOT EXISTS permissions (
    permission_id SERIAL PRIMARY KEY,
    table_id INTEGER NOT NULL,
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
    table_id,
    mode,
    label
) VALUES (
  $1,$2,$3
)
RETURNING *;

-- name: UpdatePermission :exec
UPDATE permissions
set table_id=$1,
    mode=$2,
    label=$3
WHERE permission_id = $4;

-- name: DeletePermission :exec
DELETE FROM permissions 
WHERE permission_id = $1;
