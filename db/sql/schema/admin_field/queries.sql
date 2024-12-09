-- name: GetAdminField :one
SELECT * FROM admin_fields
WHERE admin_field_id = ? LIMIT 1;

-- name: CountAdminField :one
SELECT COUNT(*)
FROM admin_fields;

-- name: GetAdminFieldId :one
SELECT admin_field_id FROM admin_fields
WHERE admin_field_id = ? LIMIT 1;

-- name: ListAdminField :many
SELECT * FROM admin_fields
ORDER BY admin_field_id;

-- name: CreateAdminField :one
INSERT INTO admin_fields (
    admin_route_id,
    parent_id,
    label,
    data,
    type,
    author,
    author_id,
    date_created,
    date_modified
    ) VALUES (
?,?,?,?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateAdminField :exec
UPDATE admin_fields
set admin_route_id = ?,
    parent_id = ?,
    label = ?,
    data = ?,
    type = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE admin_field_id = ?
    RETURNING *;

-- name: DeleteAdminField :exec
DELETE FROM admin_fields
WHERE admin_field_id = ?;

-- name: ListAdminFieldByRouteId :many
SELECT admin_field_id, admin_route_id, parent_id, label, data, type
FROM admin_fields
WHERE admin_route_id = ?;

