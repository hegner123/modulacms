-- name: UtilityGetAdminDatatypes :many
select admin_dt_id, label from admin_datatype;

-- name: UtilityGetAdminfields :many
select admin_field_id, label from admin_field;

-- name: UtilityGetAdminRoutes :many
select admin_route_id, slug from admin_route;

-- name: UtilityGetDatatypes :many
select datatype_id, label from datatype;

-- name: UtilityGetFields :many
select field_id, label from field;

-- name: UtilityGetMedia :many
select id, name from media;

-- name: UtilityGetMediaDimension :many
select id, label from media_dimension;

-- name: UtilityGetRoute :many
select route_id, slug from route;

-- name: UtilityGetTables :many
select id, label from tables;

-- name: UtilityGetToken :many
select id, user_id  from token;

-- name: UtilityGetUsers :many
select user_id, username from user;

-- name: UtilityRecordCount :many
SELECT 'admin_datatype' AS table_name, COUNT(*) AS row_count FROM admin_datatype
UNION ALL
SELECT 'admin_field' AS table_name, COUNT(*) AS row_count FROM admin_field
UNION ALL
SELECT 'admin_route' AS table_name, COUNT(*) AS row_count FROM admin_route
UNION ALL
SELECT 'datatype' AS table_name, COUNT(*) AS row_count FROM datatype
UNION ALL
SELECT 'field' AS table_name, COUNT(*) AS row_count FROM field
UNION ALL
SELECT 'route' AS table_name, COUNT(*) AS row_count FROM route
UNION ALL
SELECT 'media' AS table_name, COUNT(*) AS row_count FROM media
UNION ALL
SELECT 'media_dimension' AS table_name, COUNT(*) AS row_count FROM media_dimension
UNION ALL
SELECT 'tables' AS table_name, COUNT(*) AS row_count FROM tables
UNION ALL
SELECT 'token' AS table_name, COUNT(*) AS row_count FROM token
UNION ALL
SELECT 'user' AS table_name, COUNT(*) AS row_count FROM user;
