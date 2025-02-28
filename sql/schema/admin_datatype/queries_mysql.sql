-- name: GetAdminDatatype :one
SELECT * FROM admin_datatypes
WHERE admin_dt_id = ? 
LIMIT 1;

-- name: CountAdminDatatype :one
SELECT COUNT(*)
FROM admin_datatypes;

-- name: GetAdminDatatypeId :one
SELECT admin_dt_id FROM admin_datatypes
WHERE admin_dt_id = ? 
LIMIT 1;

-- name: ListAdminDatatype :many
SELECT * FROM admin_datatypes
ORDER BY admin_dt_id;

-- name: ListAdminDatatypeTree :many
SELECT 
    child.admin_dt_id AS child_id,
    child.label AS child_label,
    parent.admin_dt_id AS parent_id,
    parent.label AS parent_label
FROM admin_datatypes AS child
LEFT JOIN admin_datatypes AS parent 
    ON child.parent_id = parent.admin_dt_id;

-- name: GetGlobalAdminDatatypeId :one
SELECT * FROM admin_datatypes
WHERE type = 'GLOBALS' 
LIMIT 1;

-- name: ListAdminDatatypeChildren :many
SELECT * FROM admin_datatypes
WHERE parent_id = ?;

-- name: CreateAdminDatatype :exec
INSERT INTO admin_datatypes (
    admin_route_id,
    parent_id,
    label,
    type,
    author,
    author_id,
    date_created,
    date_modified,
    history
) VALUES (
    ?,?,?,?,?,?,?,?,?
);
-- To retrieve the inserted row, consider using LAST_INSERT_ID() in a subsequent SELECT.

-- name: UpdateAdminDatatype :exec
UPDATE admin_datatypes
SET admin_route_id = ?,
    parent_id = ?,
    label = ?,
    type = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
WHERE admin_dt_id = ?;
-- To retrieve the updated row, execute a subsequent SELECT.

-- name: DeleteAdminDatatype :exec
DELETE FROM admin_datatypes
WHERE admin_dt_id = ?;

-- name: ListAdminDatatypeByRouteId :many
SELECT admin_dt_id, admin_route_id, parent_id, label, type, history
FROM admin_datatypes
WHERE admin_route_id = ?;

-- name: GetRootAdminDtByAdminRtId :one
SELECT admin_dt_id, admin_route_id, parent_id, label, type, history
FROM admin_datatypes
WHERE admin_route_id = ?
ORDER BY admin_dt_id;

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

