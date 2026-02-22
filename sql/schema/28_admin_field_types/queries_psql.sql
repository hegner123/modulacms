-- name: DropAdminFieldTypesTable :exec
DROP TABLE admin_field_types;

-- name: CreateAdminFieldTypesTable :exec
CREATE TABLE IF NOT EXISTS admin_field_types (
    admin_field_type_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_field_type_id) = 26),
    type TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL
);

-- name: GetAdminFieldTypes :one
SELECT * FROM admin_field_types
WHERE admin_field_type_id = $1 LIMIT 1;

-- name: GetAdminFieldTypesByType :one
SELECT * FROM admin_field_types
WHERE type = $1 LIMIT 1;

-- name: CountAdminFieldTypes :one
SELECT COUNT(*)
FROM admin_field_types;

-- name: ListAdminFieldTypes :many
SELECT * FROM admin_field_types
ORDER BY label;

-- name: CreateAdminFieldTypes :one
INSERT INTO admin_field_types(
    admin_field_type_id,
    type,
    label
) VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: UpdateAdminFieldTypes :exec
UPDATE admin_field_types
SET type=$1,
    label=$2
WHERE admin_field_type_id = $3;

-- name: DeleteAdminFieldTypes :exec
DELETE FROM admin_field_types
WHERE admin_field_type_id = $1;
