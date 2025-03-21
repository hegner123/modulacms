

-- name: CheckAuthorIdExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE user_id=?);

-- name: CheckAuthorExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username=?);

-- name: CheckAdminRouteExists :one
SELECT EXISTS(SELECT 1 FROM admin_routes WHERE admin_route_id=?);

-- name: CheckAdminParentExists :one
SELECT EXISTS(SELECT 1 FROM admin_datatypes WHERE admin_datatype_id =?);

-- name: CheckRouteExists :one
SELECT EXISTS(SELECT 1 FROM routes WHERE route_id=?);

-- name: CheckParentExists :one
SELECT EXISTS(SELECT 1 FROM datatypes WHERE datatype_id =?);

