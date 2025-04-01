

-- name: CheckAuthorIdExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1);

-- name: CheckAuthorExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- name: CheckAdminRouteExists :one
SELECT EXISTS(SELECT 1 FROM admin_routes WHERE admin_route_id = $1);

-- name: CheckAdminParentExists :one
SELECT EXISTS(SELECT 1 FROM admin_datatypes WHERE admin_datatype_id = $1);

-- name: CheckRouteExists :one
SELECT EXISTS(SELECT 1 FROM routes WHERE route_id = $1);

-- name: CheckParentExists :one
SELECT EXISTS(SELECT 1 FROM datatypes WHERE datatype_id = $1);

