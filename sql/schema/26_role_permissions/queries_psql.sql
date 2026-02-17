-- name: CreateRolePermissionsTable :exec
CREATE TABLE IF NOT EXISTS role_permissions (
    id TEXT PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    role_id TEXT NOT NULL REFERENCES roles(role_id) ON UPDATE CASCADE ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES permissions(permission_id) ON UPDATE CASCADE ON DELETE CASCADE,
    UNIQUE(role_id, permission_id)
);

-- name: CreateRolePermissionsIndexRole :exec
CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id);

-- name: CreateRolePermissionsIndexPermission :exec
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission_id);

-- name: DropRolePermissionsTable :exec
DROP TABLE IF EXISTS role_permissions;

-- name: GetRolePermission :one
SELECT * FROM role_permissions WHERE id = $1 LIMIT 1;

-- name: ListRolePermission :many
SELECT * FROM role_permissions ORDER BY id;

-- name: ListRolePermissionByRoleID :many
SELECT * FROM role_permissions WHERE role_id = $1 ORDER BY id;

-- name: ListRolePermissionByPermissionID :many
SELECT * FROM role_permissions WHERE permission_id = $1 ORDER BY id;

-- name: ListPermissionLabelsByRoleID :many
SELECT p.label FROM role_permissions rp
JOIN permissions p ON rp.permission_id = p.permission_id
WHERE rp.role_id = $1 ORDER BY p.label;

-- name: CreateRolePermission :one
INSERT INTO role_permissions (id, role_id, permission_id) VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteRolePermission :exec
DELETE FROM role_permissions WHERE id = $1;

-- name: DeleteRolePermissionByRoleID :exec
DELETE FROM role_permissions WHERE role_id = $1;

-- name: CountRolePermission :one
SELECT COUNT(*) FROM role_permissions;

-- name: GetRoleByLabel :one
SELECT * FROM roles WHERE label = $1 LIMIT 1;

-- name: GetPermissionByLabel :one
SELECT * FROM permissions WHERE label = $1 LIMIT 1;
