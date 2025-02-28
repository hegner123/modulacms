-- name: GetAdminDatatype :one
SELECT * FROM admin_datatypes
WHERE admin_dt_id = $1
LIMIT 1;

-- name: CountAdminDatatype :one
SELECT COUNT(*)
FROM admin_datatypes;

-- name: GetAdminDatatypeId :one
SELECT admin_dt_id FROM admin_datatypes
WHERE admin_dt_id = $1
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
WHERE parent_id = $1;

-- name: CreateAdminDatatype :one
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
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateAdminDatatype :exec
UPDATE admin_datatypes
SET admin_route_id = $1,
    parent_id = $2,
    label = $3,
    type = $4,
    author = $5,
    author_id = $6,
    date_created = $7,
    date_modified = $8,
    history = $9
WHERE admin_dt_id = $10
RETURNING *;

-- name: DeleteAdminDatatype :exec
DELETE FROM admin_datatypes
WHERE admin_dt_id = $1;

-- name: ListAdminDatatypeByRouteId :many
SELECT admin_dt_id, admin_route_id, parent_id, label, type, history
FROM admin_datatypes
WHERE admin_route_id = $1;

-- name: GetRootAdminDtByAdminRtId :one
SELECT admin_dt_id, admin_route_id, parent_id, label, type, history
FROM admin_datatypes
WHERE admin_route_id = $1
ORDER BY admin_dt_id;

-- name: CheckAuthorIdExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1);

-- name: CheckAuthorExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- name: CheckAdminRouteExists :one
SELECT EXISTS(SELECT 1 FROM admin_routes WHERE admin_route_id = $1);

-- name: CheckAdminParentExists :one
SELECT EXISTS(SELECT 1 FROM admin_datatypes WHERE admin_dt_id = $1);

-- name: CheckRouteExists :one
SELECT EXISTS(SELECT 1 FROM routes WHERE route_id = $1);

-- name: CheckParentExists :one
SELECT EXISTS(SELECT 1 FROM datatypes WHERE datatype_id = $1);

