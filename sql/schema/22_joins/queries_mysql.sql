

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
