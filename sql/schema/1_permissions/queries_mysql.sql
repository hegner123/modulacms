-- name: DropPermissionTable :exec
DROP TABLE permissions;

-- name: CreatePermissionTable :exec
CREATE TABLE IF NOT EXISTS permissions (
    permission_id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NOT NULL,
    system_protected BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT perm_label_unique UNIQUE (label)
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

-- name: CreatePermission :exec
INSERT INTO permissions(
    permission_id,
    label,
    system_protected
) VALUES (
    ?,
    ?,
    ?
);

-- name: UpdatePermission :exec
UPDATE permissions
SET label=?,
    system_protected=?
WHERE permission_id = ?;

-- name: DeletePermission :exec
DELETE FROM permissions
WHERE permission_id = ?;
