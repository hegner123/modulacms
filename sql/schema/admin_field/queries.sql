-- name: GetAdminField :one
SELECT * FROM admin_field
WHERE admin_field_id = ? LIMIT 1;

-- name: CountAdminField :one
SELECT COUNT(*)
FROM admin_field;

-- name: GetAdminFieldId :one
SELECT admin_field_id FROM admin_field
WHERE admin_field_id = ? LIMIT 1;

-- name: ListAdminField :many
SELECT * FROM admin_field
ORDER BY admin_field_id;

-- name: CreateAdminField :one
INSERT INTO admin_field (
    adminrouteid,
    parentid,
    label,
    data,
    type,
    author,
    authorid,
    datecreated,
    datemodified
    ) VALUES (
?,?, ?,?, ?, ?,?, ?,?
    ) RETURNING *;


-- name: UpdateAdminField :exec
UPDATE admin_field
set adminrouteid = ?,
    parentid = ?,
    label = ?,
    data = ?,
    type = ?,
    author = ?,
    authorid = ?,
    datecreated = ?,
    datemodified = ?
    WHERE admin_field_id = ?
    RETURNING *;

-- name: DeleteAdminField :exec
DELETE FROM admin_field
WHERE admin_field_id = ?;

-- name: ListAdminFieldByRouteId :many
SELECT admin_field_id, adminrouteid, parentid, label, data, type
FROM admin_field
WHERE adminrouteid = ?;

