-- name: UtilityGetAdminDatatypes :many
select admin_dt_id, label from admin_datatypes;

-- name: UtilityGetAdminfields :many
select admin_field_id, label from admin_fields;

-- name: UtilityGetAdminRoutes :many
select admin_route_id, slug from admin_routes;

-- name: UtilityGetDatatypes :many
select datatype_id, label from datatypes;

-- name: UtilityGetFields :many
select field_id, label from fields;

-- name: UtilityGetMedia :many
select media_id, name from media;

-- name: UtilityGetMediaDimension :many
select md_id, label from media_dimensions;

-- name: UtilityGetRoute :many
select route_id, slug from routes;

-- name: UtilityGetTables :many
select id, label from tables;

-- name: UtilityGetToken :many
select id, user_id  from tokens;

-- name: UtilityGetUsers :many
select user_id, username from users;

-- name: UtilityRecordCount :many
SELECT 'admin_datatypes' AS table_name, COUNT(*) AS row_count FROM admin_datatypes
UNION ALL
SELECT 'admin_fields' AS table_name, COUNT(*) AS row_count FROM admin_fields
UNION ALL
SELECT 'admin_routes' AS table_name, COUNT(*) AS row_count FROM admin_routes
UNION ALL
SELECT 'datatypes' AS table_name, COUNT(*) AS row_count FROM datatypes
UNION ALL
SELECT 'fields' AS table_name, COUNT(*) AS row_count FROM fields
UNION ALL
SELECT 'routes' AS table_name, COUNT(*) AS row_count FROM routes
UNION ALL
SELECT 'media' AS table_name, COUNT(*) AS row_count FROM media
UNION ALL
SELECT 'media_dimensions' AS table_name, COUNT(*) AS row_count FROM media_dimensions
UNION ALL
SELECT 'tables' AS table_name, COUNT(*) AS row_count FROM tables
UNION ALL
SELECT 'tokens' AS table_name, COUNT(*) AS row_count FROM tokens
UNION ALL
SELECT 'users' AS table_name, COUNT(*) AS row_count FROM users;
