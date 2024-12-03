
-- name: GetAdminDatatype :one
SELECT * FROM admin_datatype
WHERE admin_dt_id = ? LIMIT 1;

-- name: CountAdminDatatype :one
SELECT COUNT(*)
FROM admin_datatype;

-- name: GetAdminDatatypeId :one
SELECT admin_dt_id FROM admin_datatype
WHERE admin_dt_id = ? LIMIT 1;

-- name: ListAdminDatatype :many
SELECT * FROM admin_datatype
ORDER BY admin_dt_id ;


-- name: CreateAdminDatatype :one
INSERT INTO admin_datatype (
    admin_route_id,
    parent_id,
    label,
    type,
    author,
    author_id,
    date_created,
    date_modified
    ) VALUES (
  ?, ?,?, ?,?, ?,?,?
    ) RETURNING *;


-- name: UpdateAdminDatatype :exec
UPDATE admin_datatype
set admin_route_id = ?,
    parent_id = ?,
    label = ?,
    type = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE admin_dt_id = ?
    RETURNING *;

-- name: DeleteAdminDatatype :exec
DELETE FROM admin_datatype
WHERE admin_dt_id = ?;

-- name: ListAdminDatatypeByRouteId :many
SELECT admin_dt_id, admin_route_id, parent_id, label, type
FROM admin_datatype
WHERE admin_route_id = ?;

-- name: CheckAuthorIdExists :one
SELECT EXISTS(SELECT 1 FROM user WHERE user_id=?);
-- name: CheckAuthorExists :one
SELECT EXISTS(SELECT 1 FROM user WHERE username=?);
-- name: CheckAdminRouteExists :one
SELECT EXISTS(SELECT 1 FROM admin_route WHERE admin_route_id=?);
-- name: CheckAdminParentExists :one
SELECT EXISTS(SELECT 1 FROM admin_datatype WHERE admin_dt_id =?);
-- name: CheckRouteExists :one
SELECT EXISTS(SELECT 1 FROM route WHERE route_id=?);
-- name: CheckParentExists :one
SELECT EXISTS(SELECT 1 FROM datatype WHERE datatype_id =?);
