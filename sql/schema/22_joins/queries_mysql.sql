

-- name: GetShallowTreeByRouteId :many
    SELECT cd.*, dt.label as datatype_label, dt.type as datatype_type
    FROM content_data cd
    JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
    WHERE cd.route_id = ?
    AND (cd.parent_id IS NULL OR cd.parent_id IN (
        SELECT content_data_id FROM content_data
        WHERE cd.parent_id IS NULL AND cd.route_id = ?
    ))
    ORDER BY cd.parent_id IS NULL DESC, cd.parent_id, cd.content_data_id;

-- name: GetContentTreeByRoute :many
SELECT cd.content_data_id, 
        cd.parent_id, 
        cd.first_child_id,
        cd.next_sibling_id,
        cd.prev_sibling_id,
        cd.datatype_id, 
        cd.route_id, 
        cd.author_id, 
        cd.date_created,
        cd.date_modified,
        cd.status,
       dt.label as datatype_label, dt.type as datatype_type
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id = ?
ORDER BY cd.parent_id IS NULL DESC, cd.parent_id, cd.content_data_id;

-- name: GetFieldDefinitionsByRoute :many  
SELECT DISTINCT f.field_id, f.label, f.type, df.datatype_id
FROM content_data cd
JOIN datatypes_fields df ON cd.datatype_id = df.datatype_id
JOIN fields f ON df.field_id = f.field_id  
WHERE cd.route_id = ?
ORDER BY df.datatype_id, f.field_id;

-- name: GetContentFieldsByRoute :many
SELECT cf.content_data_id, cf.field_id, cf.field_value
FROM content_data cd
JOIN content_fields cf ON cd.content_data_id = cf.content_data_id
WHERE cd.route_id = ?
ORDER BY cf.content_data_id, cf.field_id;

-- name: ListRoutesByDatatype :many
SELECT DISTINCT r.route_id, r.slug, r.title, r.status, r.author_id, r.date_created, r.date_modified
FROM routes r
INNER JOIN content_data cd ON r.route_id = cd.route_id
WHERE cd.datatype_id = ?
ORDER BY r.title;

-- name: GetRouteTreeByRouteID :many
SELECT 
    cd.content_data_id,
    cd.parent_id,
    cd.first_child_id,
    cd.next_sibling_id,
    cd.prev_sibling_id,
    dt.label AS datatype_label,
    dt.type AS datatype_type,
    f.label AS field_label,
    f.type AS field_type,
    cf.field_value
FROM content_data cd
    INNER JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
    INNER JOIN datatypes_fields df ON dt.datatype_id = df.datatype_id
    INNER JOIN fields f ON df.field_id = f.field_id
    LEFT JOIN content_fields cf ON cd.content_data_id = cf.content_data_id 
        AND f.field_id = cf.field_id
WHERE cd.route_id = ?
ORDER BY cd.content_data_id, f.field_id;

-- name: ListRootContentSummary :many
SELECT
    cd.content_data_id,
    cd.route_id,
    cd.datatype_id,
    r.slug AS route_slug,
    r.title AS route_title,
    dt.label AS datatype_label,
    cd.date_created,
    cd.date_modified
FROM content_data cd
    INNER JOIN routes r ON cd.route_id = r.route_id
    INNER JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.parent_id IS NULL
    AND dt.type = 'ROOT'
ORDER BY dt.label, r.slug;

-- name: ListAdminContentDataWithDatatypeByRoute :many
SELECT
    acd.admin_content_data_id, acd.parent_id, acd.first_child_id,
    acd.next_sibling_id, acd.prev_sibling_id, acd.admin_route_id,
    acd.admin_datatype_id, acd.author_id, acd.status,
    acd.date_created, acd.date_modified,
    adt.admin_datatype_id AS dt_admin_datatype_id,
    adt.parent_id AS dt_parent_id,
    adt.label AS dt_label,
    adt.type AS dt_type,
    adt.author_id AS dt_author_id,
    adt.date_created AS dt_date_created,
    adt.date_modified AS dt_date_modified
FROM admin_content_data acd
JOIN admin_datatypes adt ON acd.admin_datatype_id = adt.admin_datatype_id
WHERE acd.admin_route_id = ?
ORDER BY acd.parent_id IS NULL DESC, acd.parent_id, acd.admin_content_data_id;

-- name: ListAdminContentFieldsWithFieldByRoute :many
SELECT
    acf.admin_content_field_id, acf.admin_route_id,
    acf.admin_content_data_id, acf.admin_field_id,
    acf.admin_field_value, acf.author_id,
    acf.date_created, acf.date_modified,
    af.admin_field_id AS f_admin_field_id,
    af.parent_id AS f_parent_id,
    af.label AS f_label,
    af.data AS f_data,
    af.validation AS f_validation,
    af.ui_config AS f_ui_config,
    af.type AS f_type,
    af.author_id AS f_author_id,
    af.date_created AS f_date_created,
    af.date_modified AS f_date_modified
FROM admin_content_fields acf
JOIN admin_fields af ON acf.admin_field_id = af.admin_field_id
WHERE acf.admin_route_id = ?
ORDER BY acf.admin_content_data_id, acf.admin_field_id;

-- name: ListContentFieldsWithFieldByContentData :many
SELECT
    cf.content_field_id, cf.route_id,
    cf.content_data_id, cf.field_id,
    cf.field_value, cf.author_id,
    cf.date_created, cf.date_modified,
    f.field_id AS f_field_id,
    f.label AS f_label,
    f.type AS f_type
FROM content_fields cf
JOIN fields f ON cf.field_id = f.field_id
WHERE cf.content_data_id = ?
ORDER BY cf.field_id;

-- name: ListFieldsWithSortOrderByDatatypeID :many
SELECT
    df.sort_order,
    f.field_id,
    f.label,
    f.type,
    f.data,
    f.validation,
    f.ui_config
FROM datatypes_fields df
JOIN fields f ON df.field_id = f.field_id
WHERE df.datatype_id = ?
ORDER BY df.sort_order, df.id;

-- name: ListUsersWithRoleLabel :many
SELECT
    u.user_id,
    u.username,
    u.name,
    u.email,
    u.role,
    r.label AS role_label,
    u.date_created,
    u.date_modified
FROM users u
JOIN roles r ON u.role = r.role_id
ORDER BY u.username;
