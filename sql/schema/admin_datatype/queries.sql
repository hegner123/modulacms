
-- name: GetAdminDatatype :one
SELECT * FROM admin_datatypes
WHERE admin_dt_id = ? LIMIT 1;

-- name: CountAdminDatatype :one
SELECT COUNT(*)
FROM admin_datatypes;

-- name: GetAdminDatatypeId :one
SELECT admin_dt_id FROM admin_datatypes
WHERE admin_dt_id = ? LIMIT 1;

-- name: ListAdminDatatype :many
SELECT * FROM admin_datatypes
ORDER BY admin_dt_id ;

-- name: GetGlobalAdminDatatypeId :one
SELECT * FROM admin_datatypes
WHERE type = "GLOBAL" LIMIT 1;

-- name: ListAdminDatatypeChildren :many
SELECT * FROM admin_datatypes
WHERE parent_id = ?;

-- name: CreateAdminDatatype :one
INSERT INTO admin_datatypes (
    admin_route_id,
    parent_id,
    label,
    type,
    author,
    author_id,
    date_created,
    date_modified
    ) VALUES (
?,?,?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateAdminDatatype :exec
UPDATE admin_datatypes
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
DELETE FROM admin_datatypes
WHERE admin_dt_id = ?;

-- name: ListAdminDatatypeByRouteId :many
SELECT admin_dt_id, admin_route_id, parent_id, label, type
FROM admin_datatypes
WHERE admin_route_id = ?;

-- name: CheckAuthorIdExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE user_id=?);
-- name: CheckAuthorExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username=?);
-- name: CheckAdminRouteExists :one
SELECT EXISTS(SELECT 1 FROM admin_routes WHERE admin_route_id=?);
-- name: CheckAdminParentExists :one
SELECT EXISTS(SELECT 1 FROM admin_datatypes WHERE admin_dt_id =?);
-- name: CheckRouteExists :one
SELECT EXISTS(SELECT 1 FROM routes WHERE route_id=?);
-- name: CheckParentExists :one
SELECT EXISTS(SELECT 1 FROM datatypes WHERE datatype_id =?);
