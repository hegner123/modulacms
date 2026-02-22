-- name: DropAdminFieldTypesTable :exec
DROP TABLE admin_field_types;

-- name: CreateAdminFieldTypesTable :exec
CREATE TABLE IF NOT EXISTS admin_field_types (
    admin_field_type_id VARCHAR(26) PRIMARY KEY NOT NULL,
    type VARCHAR(255) NOT NULL,
    label VARCHAR(255) NOT NULL,
    CONSTRAINT admin_field_types_type_unique UNIQUE (type)
);

-- name: GetAdminFieldTypes :one
SELECT * FROM admin_field_types
WHERE admin_field_type_id = ? LIMIT 1;

-- name: GetAdminFieldTypesByType :one
SELECT * FROM admin_field_types
WHERE type = ? LIMIT 1;

-- name: CountAdminFieldTypes :one
SELECT COUNT(*)
FROM admin_field_types;

-- name: ListAdminFieldTypes :many
SELECT * FROM admin_field_types
ORDER BY label;

-- name: CreateAdminFieldTypes :exec
INSERT INTO admin_field_types(
    admin_field_type_id,
    type,
    label
) VALUES (
    ?,
    ?,
    ?
);

-- name: UpdateAdminFieldTypes :exec
UPDATE admin_field_types
SET type=?,
    label=?
WHERE admin_field_type_id = ?;

-- name: DeleteAdminFieldTypes :exec
DELETE FROM admin_field_types
WHERE admin_field_type_id = ?;
