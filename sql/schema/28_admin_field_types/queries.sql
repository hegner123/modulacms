-- name: DropAdminFieldTypeTable :exec
DROP TABLE admin_field_types;

-- name: CreateAdminFieldTypeTable :exec
CREATE TABLE IF NOT EXISTS admin_field_types (
    admin_field_type_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_field_type_id) = 26),
    type TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL
);

-- name: GetAdminFieldType :one
SELECT * FROM admin_field_types
WHERE admin_field_type_id = ? LIMIT 1;

-- name: GetAdminFieldTypesByType :one
SELECT * FROM admin_field_types
WHERE type = ? LIMIT 1;

-- name: CountAdminFieldType :one
SELECT COUNT(*)
FROM admin_field_types;

-- name: ListAdminFieldType :many
SELECT * FROM admin_field_types
ORDER BY label;

-- name: CreateAdminFieldType :one
INSERT INTO admin_field_types(
    admin_field_type_id,
    type,
    label
) VALUES (
    ?,
    ?,
    ?
)
RETURNING *;

-- name: UpdateAdminFieldType :exec
UPDATE admin_field_types
SET type=?,
    label=?
WHERE admin_field_type_id = ?;

-- name: DeleteAdminFieldType :exec
DELETE FROM admin_field_types
WHERE admin_field_type_id = ?;
