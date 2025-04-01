-- name: UtilityGetAdminDatatypes :many
SELECT admin_datatype_id, label FROM admin_datatypes;

-- name: UtilityGetAdminfields :many
SELECT admin_field_id, label FROM admin_fields;

-- name: UtilityGetAdminRoutes :many
SELECT admin_route_id, slug FROM admin_routes;

-- name: UtilityGetDatatypes :many
SELECT datatype_id, label FROM datatypes;

-- name: UtilityGetFields :many
SELECT field_id, label FROM fields;

-- name: UtilityGetMedia :many
SELECT media_id, name FROM media;

-- name: UtilityGetMediaDimension :many
SELECT md_id, label FROM media_dimensions;

-- name: UtilityGetRoute :many
SELECT route_id, slug FROM routes;

-- name: UtilityGetTables :many
SELECT id, label FROM tables;

-- name: UtilityGetToken :many
SELECT id, user_id  FROM tokens;

-- name: UtilityGetUsers :many
SELECT user_id, username FROM users;

-- name: UtilityRecordCOUNT :many
SELECT 'admin_datatypes' AS table_name, COUNT(*) AS row_COUNT FROM admin_datatypes
UNION ALL
SELECT 'admin_fields' AS table_name, COUNT(*) AS row_COUNT FROM admin_fields
UNION ALL
SELECT 'admin_routes' AS table_name, COUNT(*) AS row_COUNT FROM admin_routes
UNION ALL
SELECT 'datatypes' AS table_name, COUNT(*) AS row_COUNT FROM datatypes
UNION ALL
SELECT 'fields' AS table_name, COUNT(*) AS row_COUNT FROM fields
UNION ALL
SELECT 'routes' AS table_name, COUNT(*) AS row_COUNT FROM routes
UNION ALL
SELECT 'media' AS table_name, COUNT(*) AS row_COUNT FROM media
UNION ALL
SELECT 'media_dimensions' AS table_name, COUNT(*) AS row_COUNT FROM media_dimensions
UNION ALL
SELECT 'tables' AS table_name, COUNT(*) AS row_COUNT FROM tables
UNION ALL
SELECT 'tokens' AS table_name, COUNT(*) AS row_COUNT FROM tokens
UNION ALL
SELECT 'users' AS table_name, COUNT(*) AS row_COUNT FROM users;

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
